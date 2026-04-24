package gioutil

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/pkg/e"
)

type App struct {
	Window *app.Window

	h atomic.Pointer[AppHandler]

	ctx    context.Context
	cancel context.CancelFunc

	closed  atomic.Bool //can only be started and closed once
	started misc.ValueWithMutex[bool]

	r []Renderer

	wg sync.WaitGroup
}

func NewApp(opts ...app.Option) *App {
	a := &App{Window: NewWindowWithOptions(opts...)}
	a.Handle(nil) //will set handler to NopAppHandler
	return a
}

// Run ...
func (a *App) Run(ctx context.Context) error {
	if a.closed.Load() {
		return e.ErrClosed
	}
	a.started.Lock()
	if a.started.V {
		return e.ErrAlreadyRunning
	}
	a.ctx, a.cancel = context.WithCancel(ctx)
	defer a.cancel()

	th := material.NewTheme()
	ops := new(op.Ops)

	a.started.V = true
	a.started.Unlock() // unlock before running loop

	// handle theme as soon as mutex unlocks
	a.Handler().HandleTheme(&th)

	defer a.Window.Perform(system.ActionClose)

	// run goroutine to track for context end and performs close action
	// we need this because Event() method blocks which makes impossible to control lifecycle with
	// context unless the Event method itself supports context, but it doesn't
	a.wg.Go(func() {
		defer a.Window.Perform(system.ActionClose)
		<-a.ctx.Done()
	})

	//we don't have to call app.Main() on Windows

	a.wg.Add(1)
	return a.loop(th, ops)
}

var ErrClosedByUser = errors.New("closed by user")

func (a *App) loop(th *material.Theme, ops *op.Ops) error {
	defer a.wg.Done()
	for {
		select {
		case <-a.ctx.Done():
			return a.ctx.Err()
		default:
		}
		ev := a.Window.Event() // blocks
		if err := a.handleEvent(ev, th, ops); err != nil {
			return err
		}
		// ...
	}
}

func (a *App) handleEvent(ev event.Event, th *material.Theme, ops *op.Ops) error {
	switch ev := ev.(type) {
	case app.DestroyEvent:
		if ev.Err == nil {
			ev.Err = ErrClosedByUser
		}
		return ev.Err
	case app.ConfigEvent:
		// uhh ?
	case app.FrameEvent:
		ctx := &RendererContext{Theme: th, Ops: ops, Event: &ev}
		gtx := app.NewContext(ops, ev)

		for _, rr := range a.r {
			callRenderer(rr, ctx, gtx)
		}
		ev.Frame(ops)
	}
	return nil
}

// WithRenderer ...
func (a *App) WithRenderer(rr ...Renderer) *App {
	if a.closed.Load() || a.startedAtomic() {
		return a
	}
	a.r = append(a.r, rr...)
	return a
}

// Close ...
func (a *App) Close(cause e.CloseCause) error {
	if !a.closed.CompareAndSwap(false, true) {
		return e.ErrAlreadyClosed
	}
	if cause == nil {
		cause = e.CloseCauseExternal
	}
	if a.startedAtomic() {
		// context must be initialized if already started
		a.cancel()
	}
	a.Handler().HandleClose(cause)
	a.Window.Perform(system.ActionClose)
	// ...
	a.wg.Wait()
	return nil
}

func (a *App) Handle(h AppHandler) {
	if h == nil {
		h = NopAppHandler{}
	}
	a.h.Store(&h)
}

func (a *App) Handler() AppHandler {
	return *a.h.Load()
}

func (a *App) startedAtomic() bool {
	a.started.Lock()
	defer a.started.Unlock()
	return a.started.V
}

func callRenderer(rr Renderer, ctx *RendererContext, gtx layout.Context) {
	rr.Render(ctx)(gtx)
}
