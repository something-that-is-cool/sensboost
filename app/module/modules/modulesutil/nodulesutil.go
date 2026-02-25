package modulesutil

import "fyne.io/fyne/v2"

type Module interface {
	CreateObjects() []fyne.CanvasObject
	Disable()
}

type ModuleWithValue[T any] interface {
	Module
	Set(T) error
	Value() (T, bool)
}

type ToggleableModule = ModuleWithValue[bool]
