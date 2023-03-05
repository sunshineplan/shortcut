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

// Cmd is a shortcut command consisting of a name and arguments.
type Cmd struct {
	name string
	args []string
}

// Command returns a new Cmd with the specified name and arguments.
// If the command fails to initialize, panic with the error.
func Command(name string, arg ...string) Cmd {
	cmd := Cmd{name, arg}
	if err := cmd.cmd(context.Background()).Err; err != nil {
		panic(err)
	}
	return cmd
}

// UnmarshalJSON unmarshals the JSON representation of a Cmd.
// If the command fails to initialize, return an error.
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

// MarshalJSON returns the JSON representation of a Cmd.
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

// cmd returns an *exec.Cmd with the specified command and arguments.
// If any arguments are passed, format them into the command's arguments.
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

// Run runs the command with the given arguments.
func (c Cmd) Run(a ...any) error {
	return c.RunContext(context.Background(), a...)
}

// RunContext runs the command with the given context and arguments.
func (c Cmd) RunContext(ctx context.Context, a ...any) error {
	cmd := c.cmd(ctx, a...)
	log.Print(cmd)
	return cmd.Run()
}

// String returns the command as a string.
func (c Cmd) String() string {
	return c.cmd(context.Background()).String()
}

// Cmds is a list of commands.
type Cmds struct {
	cmds []Cmd
}

// Commands returns a new Cmds with the specified commands.
func Commands(cmd ...Cmd) Cmds {
	return Cmds{cmd}
}

// UnmarshalJSON unmarshals the JSON representation of a Cmds.
// If any command fails to initialize, return an error.
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

// MarshalJSON marshals the list of commands to JSON. If the list has only one
// command, it marshals that command directly. Otherwise, it marshals the
// list of commands.
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

// cmd processes the list of commands and returns a list of *exec.Cmd
// objects. If arguments are provided, it substitutes them into the commands.
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

// Run executes the list of commands with the given arguments.
func (c Cmds) Run(a ...any) error {
	return c.RunContext(context.Background(), a...)
}

// RunContext executes the list of commands with the given context and arguments.
func (c Cmds) RunContext(ctx context.Context, a ...any) error {
	for _, cmd := range c.cmd(ctx, a...) {
		log.Print(cmd)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// String returns the string representation of the list of commands.
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
