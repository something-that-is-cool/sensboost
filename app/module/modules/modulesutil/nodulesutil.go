package modulesutil

import (
	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

type Module interface {
	e.ErrorHandler
	CreateObjects() []fyne.CanvasObject
	Disable(cause e.ActionCause)
}

type ToggleableModule interface {
	Module
	UpdateState(v bool, cause e.ActionCause, opts ...any) error
	State() bool
}

type ModuleWithValue[T any] interface {
	Module
	SetValue(v T, cause e.ActionCause, opts ...any) error
	Value() (T, bool)
}

type ToggleableModuleWithValue[T any] interface {
	Module
	UpdateState(v bool, cause e.ActionCause, opts ...any) error
	State() bool
	SetValue(v T, cause e.ActionCause, opts ...any) error
	Value() (T, bool)
}

type SignatureSettings struct {
	Signature mem.Signature
	Patch     mem.Signature //optional for nop sig toggler
	Original  []byte        //optional
	Offset    uintptr
}
