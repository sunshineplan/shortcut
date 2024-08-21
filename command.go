package shortcut

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

var (
	_ Shortcut = Cmd{}
	_ Shortcut = Cmds{}
)

const (
	sep     = "|@$|"
	newline = "|@\n$|"
)

// Cmd is a shortcut command consisting of a name and arguments.
type Cmd struct {
	name string
	n    int
	args []string
	env  []string
}

// Command returns a new Cmd with the specified name and arguments.
// If the command fails to initialize, panic with the error.
// If n < 0, first arg with %s will be treated as a loop argument.
func Command(name string, n int, arg ...string) *Cmd {
	cmd := &Cmd{name, n, arg, nil}
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
		N    int
		Args []string
		Env  []string
	}
	if err := json.Unmarshal(b, &cmd); err != nil {
		return err
	} else {
		cmd := Cmd{cmd.Name, cmd.N, cmd.Args, cmd.Env}
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
			N    int      `json:"n"`
			Args []string `json:"args"`
			Env  []string `json:"env"`
		}{
			c.name,
			c.n,
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
	if c.n < 0 {
		index := slices.IndexFunc(c.args, func(s string) bool {
			return strings.Contains(s, "%s")
		})
		var args []string
		for i, s := range c.args {
			if i == index {
				for _, a := range a {
					args = append(args, fmt.Sprintf(s, a))
				}
			} else {
				args = append(args, s)
			}
		}
		c.args = args
	} else if len(a) != 0 {
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
	if args, l := c.Args(), len(a); args < 0 && l == 0 {
		return errors.New("loop arguments must have at least one argument")
	} else if args >= 0 && args != l {
		return badArgs(args, l)
	}
	cmd, env := c.cmd(ctx, a...)
	fmt.Println(timestamp(), cmdString(cmd, env))
	return cmd.Run()
}

func (c Cmd) Args() int {
	return c.n
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
	if l := len(c.cmds); l == 0 {
		return nil, nil
	} else if l == 1 {
		cmd, env := c.cmds[0].cmd(ctx, a...)
		cmds = append(cmds, cmd)
		envs = append(envs, env)
		return
	}
	if len(a) != 0 {
		var args string
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
	if l := len(c.cmds); l == 0 {
		return errors.New("empty commands")
	} else if l == 1 {
		return c.cmds[0].RunContext(ctx, a...)
	}
	if args := c.Args(); args == -1 {
		return errors.New("commands do not support loop arguments")
	} else if l := len(a); args != l {
		return badArgs(args, l)
	}
	cmds, envs := c.cmd(ctx, a...)
	for i, cmd := range cmds {
		fmt.Println(timestamp(), cmdString(cmd, envs[i]))
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c Cmds) Args() (n int) {
	for _, i := range c.cmds {
		if i.n < 0 {
			return -1
		}
		n += i.n
	}
	return
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

func badArgs(need, got int) error {
	return fmt.Errorf("shortcut: bad args number; need %d, got %d", need, got)
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

func timestamp() string {
	return time.Now().Format("2006/01/02 15:04:05")
}
