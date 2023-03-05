package shortcut

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
)

func TestCommand(t *testing.T) {
	cd, err := exec.LookPath("cd")
	if err != nil {
		t.Fatal(err)
	}
	a := Command("cd", "%s")
	cmd := a.cmd(context.Background(), "test")
	if cmd := cmd.String(); cmd != cd+" test" {
		t.Errorf("expected %q; got %q", cd+" test", cmd)
	}
	b := Commands(Command("cd", "%s"), Command("cd", "%s"))
	cmds := b.cmd(context.Background(), "test1", "test2")
	var res string
	for i, cmd := range cmds {
		if i != 0 {
			res += "\n"
		}
		res += cmd.String()
	}
	if expect := fmt.Sprintf("%s test1\n%[1]s test2", cd); res != expect {
		t.Errorf("expected %q; got %q", expect, res)
	}
}
