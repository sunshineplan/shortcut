package shortcut

import (
	"encoding/json"
	"os"
	"sync"
)

type Map struct {
	m sync.Map
}

func NewMap() *Map {
	return &Map{}
}

func (m *Map) Store(key Key, cmd ...Cmd) {
	switch len(cmd) {
	case 0:
		panic("no commands provided for key:" + key)
	case 1:
		m.m.Store(key, cmd[0])
	default:
		m.m.Store(key, Commands(cmd...))
	}
}

func (m *Map) Load(key Key) (Shortcut, bool) {
	if value, ok := m.m.Load(key); ok {
		return value.(Shortcut), true
	} else {
		return nil, false
	}
}

func (m *Map) Delete(key Key) {
	m.m.Delete(key)
}

func (m *Map) Range(f func(Key, Shortcut) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(Key), value.(Shortcut))
	})
}

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

func (m *Map) FromFile(file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return m.FromJSON(b)
}
