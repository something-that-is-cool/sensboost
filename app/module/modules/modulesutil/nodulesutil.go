package modulesutil

import "fyne.io/fyne/v2"

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
	Signature []byte
	Original  []byte  //optional
	Patch     []byte  //optional for non byte toggler
	Offset    uintptr //optional
}
