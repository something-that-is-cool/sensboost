package gioutil

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
)

// this thing may change slightly later

type (
	W = layout.Widget
	D = layout.Dimensions
)

type Renderer interface {
	Render(ctx *RendererContext) W
}

type RendererContext struct {
	Theme *material.Theme
	Ops   *op.Ops
	Event *app.FrameEvent
}
