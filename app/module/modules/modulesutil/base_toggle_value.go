package modulesutil

import (
	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/pkg/e"
)

type ValuedToggleableModule[T any] = interface {
	ToggleableModule
	ModuleWithValue[T]
}

func NewBaseToggleableValue[T any](m ValuedToggleableModule[T], name, desc string) module.Module {
	return &BaseToggleableValue[T]{
		ValuedToggleableModule: m,
		name:                   name,
		desc:                   desc,
	}
}

var _ module.Module = (*BaseToggleableValue[struct{}])(nil)

type BaseToggleableValue[T any] struct {
	ValuedToggleableModule[T]
	name, desc string
}

func (b *BaseToggleableValue[T]) Name() string {
	return b.name
}

func (b *BaseToggleableValue[T]) Description() string {
	return b.desc
}

func (b *BaseToggleableValue[T]) HandleError(source string, err error) {
	b.ValuedToggleableModule.HandleError(source, err)
}

func (b *BaseToggleableValue[T]) SetValue(v T, cause e.ActionCause, opts ...any) error {
	return b.ValuedToggleableModule.SetValue(v, cause, opts...)
}

func (b *BaseToggleableValue[T]) Value() (T, bool) {
	return b.ValuedToggleableModule.Value()
}

func (b *BaseToggleableValue[T]) CreateObjects() []fyne.CanvasObject {
	return b.ValuedToggleableModule.CreateObjects()
}

func (b *BaseToggleableValue[T]) Disable(cause e.ActionCause) {
	b.ValuedToggleableModule.Disable(cause)
}

func (b *BaseToggleableValue[T]) Edit(p module.Property, cause e.ActionCause) {
	b.ValuedToggleableModule.Edit(p, cause)
}

func (b *BaseToggleableValue[T]) UpdateState(v bool, cause e.ActionCause, opts ...any) error {
	return b.ValuedToggleableModule.UpdateState(v, cause, opts...)
}

func (b *BaseToggleableValue[T]) State() bool {
	return b.ValuedToggleableModule.State()
}
