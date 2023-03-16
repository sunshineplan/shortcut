package shortcut

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCommand(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	test := filepath.Join(pwd, "shortcut")
	if err := os.WriteFile(test, nil, 0640); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(test)

	a := Command(test, "%s")
	cmd := a.cmd(context.Background(), "test")
	if cmd := cmd.String(); cmd != test+" test" {
		t.Errorf("expected %q; got %q", test+" test", cmd)
	}
	b := Commands(Command(test, "%s"), Command(test, "%s"))
	cmds := b.cmd(context.Background(), "test1", "test2")
	var res string
	for i, cmd := range cmds {
		if i != 0 {
			res += "\n"
		}
		res += cmd.String()
	}
	if expect := fmt.Sprintf("%s test1\n%[1]s test2", test); res != expect {
		t.Errorf("expected %q; got %q", expect, res)
	}
	c := Command(test)
	c.Env("TEST=test")
	if expect := fmt.Sprintf("TEST=test %s", test); c.String() != expect {
		t.Errorf("expected %q; got %q", expect, res)
	}
}
