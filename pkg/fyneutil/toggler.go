package fyneutil

import (
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/pkg/e"
)

const (
	ToggleEnabled  = "enabled"
	ToggleDisabled = "disabled"
)

type Toggler struct {
	Handler      e.ErrorHandler
	Action       func(v bool, cause e.ActionCause) error
	DefaultCause e.ActionCause
	Check        *widget.Check
}

func (t *Toggler) Create() *widget.Check {
	if t.Action == nil {
		panic("must set action")
	}
	if t.Check == nil {
		t.Check = new(widget.Check)
	}
	if t.Handler == nil {
		t.Handler = e.NopErrorHandler{}
	}
	if t.DefaultCause == nil {
		t.DefaultCause = e.ActionCauseExternal
	}
	t.Check.Text = ToggleDisabled
	prev := t.Check.OnChanged

	t.Check.OnChanged = func(v bool) {
		if prev != nil {
			defer prev(v)
		}
		t.Set(v, t.DefaultCause, true)
	}
	return t.Check
}

type (
	TogglerOptionNotRefresh     struct{}
	TogglerOptionOnlyCallAction struct{}
)

func (t *Toggler) Set(v bool, cause e.ActionCause, opts ...any) {
	notRefresh, onlyCallAction := handleTogglerOpts(opts)
	if err := t.Action(v, cause); err != nil {
		t.Handler.HandleError("do action", err)
		return
	}
	if onlyCallAction {
		return
	}
	if v {
		t.Check.Text = ToggleEnabled
		t.Check.Checked = true
	} else {
		t.Check.Text = ToggleDisabled
		t.Check.Checked = false
	}
	if notRefresh {
		return
	}
	t.Check.Refresh()
}

func handleTogglerOpts(x []any) (notRefresh, onlyCallAction bool) {
	for _, v := range x {
		switch v.(type) {
		case TogglerOptionNotRefresh:
			notRefresh = true
		case TogglerOptionOnlyCallAction:
			onlyCallAction = true
		}
	}
	return
}
