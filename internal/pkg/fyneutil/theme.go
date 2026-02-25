package fyneutil

import (
	"image/color"

	"fyne.io/fyne/v2"
)

type VariantTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func NewVariantTheme(variant fyne.ThemeVariant) *VariantTheme {
	return &VariantTheme{
		Theme:   fyne.CurrentApp().Settings().Theme(),
		variant: variant,
	}
}

// Color ...
func (t *VariantTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, t.variant)
}
