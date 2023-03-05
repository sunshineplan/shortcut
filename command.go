package shortcut

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	_ Shortcut = Cmd{}
	_ Shortcut = Cmds{}
)

type Cmd struct {
	name string
	args []string
}

func Command(name string, arg ...string) Cmd {
	cmd := Cmd{name, arg}
	if err := cmd.cmd(context.Background()).Err; err != nil {
		panic(err)
	}
	return cmd
}

func (c *Cmd) UnmarshalJSON(b []byte) error {
	var cmd struct {
		Name string
		Args []string
	}
	if err := json.Unmarshal(b, &cmd); err != nil {
		return err
	} else {
		cmd := Cmd{cmd.Name, cmd.Args}
		if err := cmd.cmd(context.Background()).Err; err != nil {
			return err
		}
		*c = cmd
		return nil
	}
}

func (c Cmd) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Name string   `json:"name"`
			Args []string `json:"args"`
		}{
			c.name,
			c.args,
		},
	)
}

func (c Cmd) cmd(ctx context.Context, a ...any) *exec.Cmd {
	if len(a) != 0 {
		const sep = "|@$|"
		args := strings.Join(c.args, sep)
		args = fmt.Sprintf(args, a...)
		c.args = strings.Split(args, sep)
	}
	cmd := exec.CommandContext(ctx, c.name, c.args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (c Cmd) Run(a ...any) error {
	return c.RunContext(context.Background(), a...)
}

func (c Cmd) RunContext(ctx context.Context, a ...any) error {
	cmd := c.cmd(ctx, a...)
	log.Print(cmd)
	return cmd.Run()
}

func (c Cmd) String() string {
	return c.cmd(context.Background()).String()
}

type Cmds struct {
	cmds []Cmd
}

func Commands(cmd ...Cmd) Cmds {
	return Cmds{cmd}
}

func (c *Cmds) UnmarshalJSON(b []byte) error {
	var cmd Cmd
	if e1 := json.Unmarshal(b, &cmd); e1 == nil {
		*c = Cmds{[]Cmd{cmd}}
	} else {
		var cmds []Cmd
		if e2 := json.Unmarshal(b, &cmds); e2 != nil {
			return errors.Join(e1, e2)
		}
		if len(cmds) == 0 {
			return errors.New("empty commands")
		}
		*c = Cmds{cmds}
	}
	return nil
}

func (c Cmds) MarshalJSON() ([]byte, error) {
	switch len(c.cmds) {
	case 0:
		return nil, errors.New("empty commands")
	case 1:
		return json.Marshal(c.cmds[0])
	default:
		return json.Marshal(c.cmds)
	}
}

func (c Cmds) cmd(ctx context.Context, a ...any) (cmds []*exec.Cmd) {
	if len(a) != 0 {
		var args string
		const sep, newline = "|@$|", "|@\n$|"
		for i, cmd := range c.cmds {
			if i != 0 {
				args += newline
			}
			args += strings.Join(cmd.args, sep)
		}
		args = fmt.Sprintf(args, a...)
		for i, cmd := range strings.Split(args, newline) {
			c.cmds[i].args = strings.Split(cmd, sep)
			cmds = append(cmds, c.cmds[i].cmd(ctx))
		}
	} else {
		for _, cmd := range c.cmds {
			cmds = append(cmds, cmd.cmd(ctx))
		}
	}
	return
}

func (c Cmds) Run(a ...any) error {
	return c.RunContext(context.Background(), a...)
}

func (c Cmds) RunContext(ctx context.Context, a ...any) error {
	for _, cmd := range c.cmd(ctx, a...) {
		log.Print(cmd)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c Cmds) String() string {
	var b strings.Builder
	for i, cmd := range c.cmds {
		if i != 0 {
			b.WriteRune('\n')
		}
		b.WriteString(cmd.String())
	}
	return b.String()
}
