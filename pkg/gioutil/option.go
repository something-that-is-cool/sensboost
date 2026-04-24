package gioutil

import (
	"gioui.org/app"
	"gioui.org/unit"
)

func StaticSize(w, h unit.Dp) app.Option {
	return func(metric unit.Metric, config *app.Config) {
		app.MinSize(w, h)(metric, config)
		app.MaxSize(w, h)(metric, config)
		// gio will automatically make window unresizable if min and max sizes are same
	}
}
