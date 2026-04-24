// SPDX-License-Identifier: Unlicense OR MIT

package main

// A simple Gio program. See https://gioui.org for more information.

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"os/signal"
	"syscall"
	"time"

	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/something-that-is-cool/zutil/pkg/gioutil"
)

// TODO: refactor current zutil/app

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := gioutil.NewApp(gioutil.StaticSize(400, 500))
	app.WithRenderer(LabelRenderer{})

	go func() {
		<-time.After(time.Second * 3)
		if err := app.Close(nil); err != nil {
			fmt.Printf("error closing app: %v\n", err)
		}
	}()

	fmt.Println("running app...")
	if err := app.Run(ctx); err != nil && (!errors.Is(err, gioutil.ErrClosedByUser) && !errors.Is(err, context.Canceled)) {
		panic(fmt.Errorf("run app: %w", err))
	}
}

type LabelRenderer struct{}

func (LabelRenderer) Render(ctx *gioutil.RendererContext) gioutil.W {
	l := material.H1(ctx.Theme, "Hello, Gio")
	maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
	l.Color = maroon
	l.Alignment = text.Middle
	return l.Layout
}
