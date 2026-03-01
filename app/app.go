package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

const (
	Name = "zutil"
	ID   = "monster.zov.zutil"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	conf Config

	wg sync.WaitGroup

	tr *win.ProcessTracker
	hm *win.HotkeyManager

	closed atomic.Bool

	userConf misc.ValueWithMutex[*UserConfig]

	data misc.ValueWithMutex[struct {
		started, init bool

		win fyne.Window
		app fyne.App

		modules *modulesMap
	}]
}

func (app *App) initUnsafe(proc *win.Process) (err error) {
	if app.data.V.init {
		return errors.New("already initialized")
	}
	app.data.V.init = true // can only initialize once
	a, w := app.deployFyneUnsafe()

	app.userConf.Lock()
	app.userConf.V, err = app.loadUserConfigUnsafe(a)
	if err != nil {
		app.userConf.Unlock()
		return fmt.Errorf("load user config: %w", err)
	}
	app.syncThemeUnsafe(app.userConf.V)
	app.userConf.Unlock()

	configs := app.setupModules(proc)
	if len(configs) == 0 {
		return errors.New("no modules created")
	}
	modules := app.createModulesFromConfigs(configs)
	app.data.V.modules = modules

	app.userConf.Lock()
	app.hm = win.HotkeyManagerConfig{Handlers: app.loadBinds(app.userConf.V)}.New()
	app.userConf.Unlock()

	c, err := app.createContent(modules, w)
	if err != nil {
		return fmt.Errorf("create content: %w", err)
	}
	app.data.V.win.SetContent(c)
	return nil
}

const windowWidth, windowHeight = 450, 550

func (app *App) deployFyneUnsafe() (fyne.App, fyne.Window) {
	app.data.V.app = fyneapp.NewWithID(ID)
	app.data.V.win = app.data.V.app.NewWindow(Name)

	app.data.V.win.SetMaster()
	app.data.V.win.CenterOnScreen()
	app.data.V.win.Resize(fyne.NewSize(windowWidth, windowHeight))
	app.data.V.win.SetFixedSize(true)
	return app.data.V.app, app.data.V.win
}

var ErrAppClosed = errors.New("app closed")

var ErrAlreadyRunning = errors.New("app is already running")

// Run ...
func (app *App) Run() error {
	if app.closed.Load() {
		return ErrAppClosed
	}
	w, err := app.doInit()
	if err != nil {
		return err
	}
	go func() {
		<-app.ctx.Done()
		if err := app.close(false, closeCauseContextClosed); err != nil && !errors.Is(err, ErrAppClosed) {
			app.conf.Logger.Error("close app", "err", err.Error())
		}
	}()
	go func() {
		defer app.tr.Close()
		app.conf.Logger.Info("running process tracker...")

		if err := app.tr.Run(app.ctx); err != nil && !errors.Is(err, context.Canceled) {
			app.conf.Logger.Error("process tracker cannot run", "err", err.Error())
			return
		}
		app.conf.Logger.Info("process tracker ended gracefully.")
	}()
	app.wg.Go(func() {
		defer app.doClose("hotkey manager", app.hm.Close)
		app.conf.Logger.Info("running hotkey manager...")

		if err := app.hm.Run(app.ctx); err != nil && !errors.Is(err, context.Canceled) {
			app.conf.Logger.Error("hotkey manager cannot run", "err", err.Error())
			return
		}
		app.conf.Logger.Info("hotkey manager ended gracefully.")
	})
	app.conf.Logger.Info("running window...")
	app.setStarted()
	w.ShowAndRun() // blocks
	// ...
	app.conf.Logger.Info("window closed.")
	return nil
}

func (app *App) doInit() (fyne.Window, error) {
	app.data.Lock()
	defer app.data.Unlock()

	if app.data.V.started {
		return nil, ErrAlreadyRunning
	}
	app.conf.Logger.Debug("initializing...")
	if err := app.initUnsafe(app.tr.Process()); err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}
	app.conf.Logger.Debug("initialized.")
	return app.data.V.win, nil
}

func (app *App) setStarted() {
	app.data.Lock()
	defer app.data.Unlock()
	app.data.V.started = true
}

// Close implements io.Closer.
func (app *App) Close() error {
	return app.close(false, nil)
}

func (app *App) CloseMain() error {
	return app.close(true, nil)
}

func (app *App) close(main bool, cause error) error {
	if !app.closed.CompareAndSwap(false, true) {
		return ErrAppClosed
	}
	start := time.Now()
	if cause == nil {
		cause = closeCauseExternal
	}
	app.conf.Logger.Info("closing app...")
	defer func() {
		app.conf.Logger.Info("closed app.", "elapsed", time.Since(start).String())
	}()
	app.closeIfStarted(cause)
	// wait for additional things to end
	app.conf.Logger.Debug("waiting for waitgroup end...")
	app.wg.Wait()
	app.conf.Logger.Debug("waitgroup ended.")
	app.conf.Logger.Debug("closing window...")
	// close window last of all !!!
	if !main {
		//fixme:hack that prevents random window deadlock on close
		fyne.Do(app.closeWin) // fyne.DoAndWait
		return nil
	}
	app.closeWin()
	return nil
}

func (app *App) closeIfStarted(cause error) {
	app.data.Lock()
	defer app.data.Unlock() // release lock earlier because we don't want to lock until wg done

	if app.data.V.started {
		// save the user config
		app.saveUserConfig(app.data.V.app)
		// if closed because parent process closed, we don't need to disable
		// all modules as it will not affect
		if !errors.Is(cause, closeCauseTrackerClosed) {
			// disable all modules before canceling context
			// if we will not do this any logic that takes our context can end
			// earlier so it won't disable modules properly
			app.disableModulesUnsafe()
		}
		app.doClose("hotkey manager", app.hm.Close)
	}
	// after we disabled all modules we can close the context safely
	app.cancel()
	app.tr.Close()
}

func (app *App) disableModulesUnsafe() {
	app.conf.Logger.Debug("disabling modules...")
	defer app.conf.Logger.Debug("disabled modules.")

	for _, m := range app.data.V.modules.AllFromFront() {
		m.Disable()
		app.conf.Logger.Debug("disabled module.", "module", m.Name())
	}
}

func (app *App) closeWin() {
	app.data.Lock()
	defer app.data.Unlock()

	if w := app.data.V.win; w != nil {
		w.Close()
		app.conf.Logger.Debug("closed window.")
		return
	}
	// make sure app.win is never set to nil after initialized
	app.conf.Logger.Error("tried to close window before it was initialized (nil)")
}

//fixme very rarely the app process can freeze forever (while the fyne window closes)
//upd: it starts closing window but not ends, so the problem is probably fyne.DoAndWait
