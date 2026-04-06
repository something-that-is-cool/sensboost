package modulesutil

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/pkg/asm"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
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

type PatchFunc = func(Settings) mem.Signature

func PatchFuncExtend(x ...byte) PatchFunc {
	return func(s Settings) mem.Signature {
		return s.Signature.ExtendFirst(x)
	}
}

func PatchFuncExtendNop(n int) PatchFunc {
	return PatchFuncExtend(mem.NopBytes(n)...)
}

func PatchFuncExtendBuilder(b *asm.Builder) PatchFunc {
	return func(s Settings) mem.Signature {
		sig := b.BuildSignature()
		b.ClearAndReturn() // the builder must no longer be used
		return s.Signature.ExtendFirst(sig.Data)
	}
}

func disableOnlyAction(t ToggleableModule, cause e.ActionCause) error {
	return t.UpdateState(false, cause, fyneutil.TogglerOptionOnlyCallAction{})
}

func extendPatchFunc(conf *Settings) bool {
	if conf.Patch.Data != nil {
		return true
	}
	if conf.PatchFunc == nil {
		return false
	}
	conf.Patch = conf.PatchFunc(*conf)
	return conf.Patch.Data != nil
}
