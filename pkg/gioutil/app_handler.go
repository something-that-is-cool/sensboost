package gioutil

import (
	"gioui.org/io/event"
	"gioui.org/widget/material"
	"github.com/something-that-is-cool/zutil/pkg/e"
)

type AppHandler interface {
	HandleEvent(ev event.Event) error
	HandleClose(cc e.CloseCause)
	// HandleTheme handles creation of theme and allows to initialize it.
	HandleTheme(th **material.Theme)
}

var _ AppHandler = (*NopAppHandler)(nil)

type NopAppHandler struct{}

func (NopAppHandler) HandleEvent(event.Event) error { return nil }
func (NopAppHandler) HandleClose(e.CloseCause)      {}
func (NopAppHandler) HandleTheme(**material.Theme)  {}
