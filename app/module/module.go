package module

import "fyne.io/fyne/v2"

type Module interface {
	Name() string
	Description() string
	CreateObjects() []fyne.CanvasObject
	Edit(Property)
	Disable()
}

type Config interface {
	Create(Property) (Module, error)
	DefaultProperty() Property
	Identifier() string // identify in config
}

type Property struct {
	Enabled bool `json:"enabled,omitempty"`
	Value   any  `json:"value,omitempty"`
}
