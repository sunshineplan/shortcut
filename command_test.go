package shortcut

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestCommand(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	a := Command(self, 1, "%s")
	cmd, _ := a.cmd(context.Background(), "test")
	if cmd := cmd.String(); cmd != self+" test" {
		t.Errorf("expected %q; got %q", self+" test", cmd)
	}
	b := Commands(Command(self, 1, "%s"), Command(self, 1, "%s"))
	cmds, _ := b.cmd(context.Background(), "test1", "test2")
	var res string
	for i, cmd := range cmds {
		if i != 0 {
			res += "\n"
		}
		res += cmd.String()
	}
	if expect := fmt.Sprintf("%s test1\n%[1]s test2", self); res != expect {
		t.Errorf("expected %q; got %q", expect, res)
	}
	c := Command(self, 0)
	c.Env("TEST=test")
	if expect := fmt.Sprintf("TEST=test %s", self); c.String() != expect {
		t.Errorf("expected %q; got %q", expect, res)
	}
}

func TestLoopCommand(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	a := Command(self, -1, "-%s")
	cmd, _ := a.cmd(context.Background(), "a", "b", "c")
	if cmd, expect := cmd.String(), self+" -a -b -c"; cmd != expect {
		t.Errorf("expected %q; got %q", expect, cmd)
	}
	b := Commands(Command(self, -1, "%s"), Command(self, 1, "%s"))
	if args := b.Args(); args != -1 {
		t.Errorf("expected -1; got %d", args)
	}
	if err := b.Run(); err == nil || err.Error() != "commands do not support loop arguments" {
		t.Errorf("expected error; got %v", err)
	}
}
