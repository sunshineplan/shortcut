package shortcut

import "context"

type Key string

type Shortcut interface {
	Run(...any) error
	RunContext(context.Context, ...any) error
	String() string
}
