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
	env  []string
}

// Command returns a new Cmd with the specified name and arguments.
// If the command fails to initialize, panic with the error.
func Command(name string, arg ...string) *Cmd {
	cmd := &Cmd{name, arg, nil}
	if err := cmd.test(); err != nil {
		panic(err)
	}
	return cmd
}

// Env sets the environment variables for the command.
// The environment variables are specified as a slice of strings,
// with each string being in the form of "key=value".
// If multiple values are specified for the same key,
// the last one takes precedence.
// If env is empty, the environment variables of the parent process are used.
func (c *Cmd) Env(env ...string) {
	c.env = env
}

// UnmarshalJSON unmarshals the JSON representation of a Cmd.
// If the command fails to initialize, return an error.
func (c *Cmd) UnmarshalJSON(b []byte) error {
	var cmd struct {
		Name string
		Args []string
		Env  []string
	}
	if err := json.Unmarshal(b, &cmd); err != nil {
		return err
	} else {
		cmd := Cmd{cmd.Name, cmd.Args, cmd.Env}
		if err := cmd.test(); err != nil {
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
			Env  []string `json:"env"`
		}{
			c.name,
			c.args,
			c.env,
		},
	)
}

func (c Cmd) test() error {
	cmd := exec.Command(c.name, c.args...)
	return cmd.Err
}

// cmd returns an *exec.Cmd with the specified command and arguments.
// If any arguments are passed, format them into the command's arguments.
func (c Cmd) cmd(ctx context.Context, a ...any) (*exec.Cmd, []string) {
	if len(a) != 0 {
		const sep = "|@$|"
		args := strings.Join(c.args, sep)
		args = fmt.Sprintf(args, a...)
		c.args = strings.Split(args, sep)
	}
	cmd := exec.CommandContext(ctx, c.name, c.args...)
	if len(c.env) != 0 {
		cmd.Env = append(os.Environ(), c.env...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, c.env
}

// Run runs the command with the given arguments.
func (c Cmd) Run(a ...any) error {
	return c.RunContext(context.Background(), a...)
}

// RunContext runs the command with the given context and arguments.
func (c Cmd) RunContext(ctx context.Context, a ...any) error {
	cmd, env := c.cmd(ctx, a...)
	log.Print(cmdString(cmd, env))
	return cmd.Run()
}

// String returns the command as a string.
func (c Cmd) String() string {
	return cmdString(c.cmd(context.Background()))
}

// Cmds is a list of commands.
type Cmds struct {
	cmds []*Cmd
}

// Commands returns a new Cmds with the specified commands.
func Commands(cmd ...*Cmd) Cmds {
	return Cmds{cmd}
}

// UnmarshalJSON unmarshals the JSON representation of a Cmds.
// If any command fails to initialize, return an error.
func (c *Cmds) UnmarshalJSON(b []byte) error {
	var cmd Cmd
	if e1 := json.Unmarshal(b, &cmd); e1 == nil {
		*c = Cmds{[]*Cmd{&cmd}}
	} else {
		var cmds []*Cmd
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
func (c Cmds) cmd(ctx context.Context, a ...any) (cmds []*exec.Cmd, envs [][]string) {
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
			cmd, env := c.cmds[i].cmd(ctx)
			cmds = append(cmds, cmd)
			envs = append(envs, env)
		}
	} else {
		for _, cmd := range c.cmds {
			cmd, env := cmd.cmd(ctx)
			cmds = append(cmds, cmd)
			envs = append(envs, env)
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
	cmds, envs := c.cmd(ctx, a...)
	for i, cmd := range cmds {
		log.Print(cmdString(cmd, envs[i]))
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

func cmdString(cmd *exec.Cmd, env []string) string {
	var b strings.Builder
	if len(env) != 0 {
		for _, i := range env {
			b.WriteString(i)
			b.WriteRune(' ')
		}
	}
	b.WriteString(cmd.String())
	return b.String()
}
