package modulesutil

import (
	"fmt"

	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/app/module"
)

const (
	ToggleEnabled  = "enabled"
	ToggleDisabled = "disabled"
)

func CheckSet(onError func(error), check *widget.Check, act func(bool, *widget.Check) error) func(bool) {
	return func(b bool) {
		if err := act(b, check); err != nil {
			onError(fmt.Errorf("do action: %w", err))
			return
		}
		if b {
			check.Text = ToggleEnabled
			check.Checked = true
		} else {
			check.Text = ToggleDisabled
			check.Checked = false
		}
		check.Refresh()
	}
}

type DefaultDisabled struct{}

func (DefaultDisabled) DefaultProperty() module.Property {
	return module.Property{Enabled: false}
}
