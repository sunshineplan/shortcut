package shortcut

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/sunshineplan/utils/choice"
)

// Map is a shortcut map that associates Key with one or multiple Cmds.
type Map struct {
	m sync.Map
}

// NewMap creates and returns a new instance of Map.
func NewMap() *Map {
	return &Map{}
}

// Store stores a Key with one or multiple Cmds in the Map.
// If no command is provided, it panics.
// If only one command is provided, it is stored directly.
// If multiple commands are provided, they are wrapped as a Cmds object and stored.
func (m *Map) Store(key Key, cmd ...*Cmd) {
	switch len(cmd) {
	case 0:
		panic("no commands provided for key:" + key)
	case 1:
		m.m.Store(key, cmd[0])
	default:
		m.m.Store(key, Commands(cmd...))
	}
}

// Load retrieves a Shortcut with the given key.
// It returns the Shortcut and true if the key exists.
// If the key does not exist, it returns nil and false.
func (m *Map) Load(key Key) (Shortcut, bool) {
	if value, ok := m.m.Load(key); ok {
		return value.(Shortcut), true
	} else {
		return nil, false
	}
}

// Delete removes the Shortcut with the given key from the Map.
func (m *Map) Delete(key Key) {
	m.m.Delete(key)
}

// Range calls the given function for each key-value pair in the Map until the function returns false.
func (m *Map) Range(f func(Key, Shortcut) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(Key), value.(Shortcut))
	})
}

// FromJSON unmarshals a JSON byte array and populates the Map with the key-value pairs.
func (m *Map) FromJSON(b []byte) error {
	var v map[Key]Cmds
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	for k, v := range v {
		m.m.Store(k, v)
	}
	return nil
}

// FromFile reads the content of the JSON file and populates the Map with the key-value pairs.
func (m *Map) FromFile(file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return m.FromJSON(b)
}

type pair struct {
	key      Key
	shortcut Shortcut
	length   int
}

func (pair pair) String() string {
	return fmt.Sprintf("%s  %s  %s", pair.key, strings.Repeat(" ", pair.length-len(pair.key)), pair.shortcut)
}

func (m *Map) pairs() (pairs []*pair) {
	var maxKeyLength int
	m.Range(func(k Key, s Shortcut) bool {
		if l := len(k); l > maxKeyLength {
			maxKeyLength = l
		}
		pairs = append(pairs, &pair{k, s, 0})
		return true
	})
	for _, i := range pairs {
		i.length = maxKeyLength
	}
	slices.SortStableFunc(pairs, func(a, b *pair) int { return cmp.Compare(a.key, b.key) })
	return
}

func (m *Map) Count() int {
	return len(m.pairs())
}

func (m *Map) Index(index int) (Key, Shortcut, error) {
	if n := m.Count(); n == 0 {
		return "", nil, fmt.Errorf("empty shortcut map")
	} else if index < 1 || index > n {
		return "", nil, fmt.Errorf("out of range(1-%d): %d", n, index)
	}
	pair := m.pairs()[index-1]
	return pair.key, pair.shortcut, nil
}

func (m *Map) Menu(showQuit bool) string {
	return choice.Menu(m.pairs(), showQuit)
}

func (m *Map) Choose() (bool, Key, Shortcut, error) {
	choice, v, err := choice.Choose(m.pairs())
	if err != nil {
		return false, "", nil, err
	} else if !choice {
		return false, "", nil, nil
	}
	return true, v.key, v.shortcut, nil
}
