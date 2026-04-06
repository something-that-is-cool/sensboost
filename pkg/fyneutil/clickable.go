package fyneutil

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type ClickableIcon struct {
	*widget.Icon
	tap func()
}

func NewClickableIcon(icon fyne.Resource, tap func()) *ClickableIcon {
	if tap == nil {
		tap = func() {}
	}
	return &ClickableIcon{Icon: widget.NewIcon(icon), tap: tap}
}

func (c *ClickableIcon) Tapped(*fyne.PointEvent) {
	c.tap()
}
