package gioutil

import "gioui.org/app"

func NewWindow() *app.Window {
	return NewWindowWithOptions()
}

func NewWindowWithOptions(opts ...app.Option) *app.Window {
	w := new(app.Window)
	if len(opts) > 0 {
		w.Option(opts...)
	}
	return w
}
