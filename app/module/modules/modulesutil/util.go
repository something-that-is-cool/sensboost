package modulesutil

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
)

type DefaultDisabled struct{}

func (DefaultDisabled) DefaultProperty() module.Property {
	return module.Property{Enabled: false}
}

var _ e.ErrorHandler = (*errorHandler)(nil)

type errorHandler struct {
	err func(error)
}

func (h errorHandler) HandleError(src string, err error) {
	if err == nil {
		return
	}
	if src == "" {
		h.err(err)
		return
	}
	h.err(fmt.Errorf("%s: %w", src, err))
}

func SyncState(m ToggleableModule, p module.Property, cause e.ActionCause) {
	m.HandleError("update state by property", m.UpdateState(p.Enabled, cause))
}

func SyncValue[T any](m ModuleWithValue[T], p module.Property, cause e.ActionCause) {
	if v, ok := p.Value.(T); ok {
		m.HandleError("update value by property", m.SetValue(v, cause))
	}
}

func disableOnlyAction(t ToggleableModule, cause e.ActionCause) error {
	return t.UpdateState(false, cause, fyneutil.TogglerOptionOnlyCallAction{})
}
