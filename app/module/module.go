package module

import "fyne.io/fyne/v2"

type Module interface {
	Name() string
	Identifier() string // identify in config
	Description() string
	CreateObjects() []fyne.CanvasObject
	Disable()
}

type Config interface {
	Create() Module
}
