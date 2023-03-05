package shortcut

import "context"

type Key string

// Shortcut is an interface for defining a shortcut command.
type Shortcut interface {
	Run(...any) error
	RunContext(context.Context, ...any) error
	String() string
}
