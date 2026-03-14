package modulesutil

import (
	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

type Module interface {
	CreateObjects() []fyne.CanvasObject
	Disable()
}

type ToggleableModule interface {
	Module
	UpdateState(bool) error
	State() bool
}

type ModuleWithValue[T any] interface {
	Module
	SetValue(T) error
	Value() (T, bool)
}

type ToggleableModuleWithValue[T any] interface {
	ToggleableModule
	ModuleWithValue[T]
}

type PointerSettings struct {
	BaseAddress uintptr
	Offsets     []uintptr
}

type SignatureSettings struct {
	Signature mem.Signature
	Patch     mem.Signature //optional for nop sig toggler
	Original  []byte        //optional
	Offset    uintptr
}
