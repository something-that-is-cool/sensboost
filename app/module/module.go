package module

import "fyne.io/fyne/v2"

type Module interface {
	Name() string
	Description() string
	CreateObjects() []fyne.CanvasObject
	Disable()
}
