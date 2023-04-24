package shortcut

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMap(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	test := filepath.Join(pwd, "shortcut")
	if err := os.WriteFile(test, nil, 0640); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(test)

	json := fmt.Sprintf(`
{
  "a": {
    "name": %q,
    "args": [
      ".."
    ]
  },
  "b": [
    {
      "name": %[1]q,
      "args": [
        ".."
      ]
    },
    {
      "name": %[1]q,
      "args": [
        "test"
      ],
	  "env": [
        "TEST=test"
      ]
    }
  ],
  "c": {
    "name": %[1]q,
    "args": [
      "%%s"
    ]
  },
  "d": [
    {
      "name": %[1]q,
      "args": [
        "%%s"
      ]
    },
    {
      "name": %[1]q,
      "args": [
        "%%s"
      ]
    }
  ],
  "e": {
    "name": %[1]q,
    "env": [
      "TEST=test"
    ]
  }
}`, test)
	m := NewMap()
	m.Store("f", Command(test, 0))
	if err := m.FromJSON([]byte(json)); err != nil {
		t.Fatal(err)
	}
	for _, testcase := range []struct {
		key Key
		cmd string
	}{
		{"a", test + " .."},
		{"b", fmt.Sprintf("%s ..\nTEST=test %[1]s test", test)},
		{"c", test + " %s"},
		{"d", fmt.Sprintf("%s %%s\n%[1]s %%s", test)},
		{"e", "TEST=test " + test},
		{"f", test},
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
