package shortcut

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestMap(t *testing.T) {
	cd, err := exec.LookPath("cd")
	if err != nil {
		cd = "cd"
	}
	echo, err := exec.LookPath("echo")
	if err != nil {
		echo = "echo"
	}

	json := `
{
  "a": {
    "name": "cd",
    "args": [
      ".."
    ]
  },
  "b": [
    {
      "name": "cd",
      "args": [
        ".."
      ]
    },
    {
      "name": "echo",
      "args": [
        "test"
      ]
    }
  ],
  "c": {
    "name": "cd",
    "args": [
      "%s"
    ]
  },
  "d": [
    {
      "name": "cd",
      "args": [
        "%s"
      ]
    },
    {
      "name": "echo",
      "args": [
        "%s"
      ]
    }
  ]
}`
	m := NewMap()
	m.Store("e", Command("cd"))
	if err := m.FromJSON([]byte(json)); err != nil {
		t.Fatal(err)
	}
	for _, testcase := range []struct {
		key Key
		cmd string
	}{
		{"a", cd + " .."},
		{"b", fmt.Sprintf("%s ..\n%s test", cd, echo)},
		{"c", cd + " %s"},
		{"d", fmt.Sprintf("%s %%s\n%s %%s", cd, echo)},
		{"e", cd},
	} {
		if sc, ok := m.Load(testcase.key); ok {
			if cmd := sc.String(); cmd != testcase.cmd {
				t.Errorf("key %s expected %q; got %q", testcase.key, testcase.cmd, cmd)
			}
		} else {
			t.Errorf("key %s not found", testcase.key)
		}
	}
}

func TestJSON(t *testing.T) {
	m := NewMap()
	if err := m.FromJSON([]byte("json")); err == nil {
		t.Error("expected error; got nil")
	}
	if err := m.FromJSON([]byte("{}")); err != nil {
		t.Errorf("expected nil; got %s", err)
	}
	if err := m.FromJSON([]byte(`{"a":{"name":1}}`)); err == nil {
		t.Error("expected error; got nil")
	}
	if err := m.FromJSON([]byte(`{"a":[]}`)); err == nil {
		t.Error("expected error; got nil")
	}
}
