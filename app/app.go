package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/elliotchance/orderedmap/v3"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

const Name = "zutil"

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	conf Config

	wg sync.WaitGroup

	tr *win.ProcessTracker

	closed, started atomic.Bool

	uConf   *UserConfig
	uConfMu sync.Mutex

	data struct {
		sync.Mutex
		modules *orderedmap.OrderedMap[module.Config, module.Module]

		win fyne.Window
		app fyne.App
	}
}

func (app *App) init(proc *win.Process) (err error) {
	app.data.Lock()
	defer app.data.Unlock()
	app.deployFyne()

	app.uConf, err = app.loadUserConfigUnsafe()
	if err != nil {
		return fmt.Errorf("load user config: %w", err)
	}
	app.uConfMu.Lock()
	app.syncThemeUnsafe(app.uConf)
	app.uConfMu.Unlock()

	configs := app.setupModules(proc)
	if len(configs) == 0 {
		return errors.New("no modules created")
	}
	modules := app.createModulesFromConfigs(configs)
	app.data.modules = modules

	c, err := app.createContent(modules)
	if err != nil {
		return fmt.Errorf("create content: %w", err)
	}
	app.data.win.SetContent(c)
	return nil
}

const windowWidth, windowHeight = 550, 550

func (app *App) deployFyne() {
	app.data.app = fyneapp.New()
	app.data.win = app.data.app.NewWindow(Name)

	app.data.win.SetMaster()
	app.data.win.CenterOnScreen()
	app.data.win.Resize(fyne.NewSize(windowWidth, windowHeight))
	app.data.win.SetFixedSize(true)
}

var ErrAppClosed = errors.New("app closed")

var ErrAlreadyRunning = errors.New("app is already running")

// Run ...
func (app *App) Run() error {
	if app.closed.Load() {
		return ErrAppClosed
	}
	if !app.started.CompareAndSwap(false, true) {
		return ErrAlreadyRunning
	}
	app.conf.Logger.Debug("initializing...")
	if err := app.init(app.tr.Process()); err != nil {
		return fmt.Errorf("init: %w", err)
	}
	app.conf.Logger.Debug("initialized.")
	go func() {
		<-app.ctx.Done()
		if err := app.Close(false); err != nil && !errors.Is(err, ErrAppClosed) {
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
	app.conf.Logger.Info("running window...")

	app.data.Lock()
	w := app.data.win
	app.data.Unlock()
	w.ShowAndRun()

	app.conf.Logger.Info("window closed.")
	return nil
}

// Close ...
func (app *App) Close(main bool) error {
	if !app.closed.CompareAndSwap(false, true) {
		return ErrAppClosed
	}
	app.conf.Logger.Info("closing app...")
	defer app.conf.Logger.Info("closed app.")

	// in case of panic do this pattern to defer unlock
	func() {
		app.conf.Logger.Debug("disabling modules...")
		defer app.conf.Logger.Debug("disabled modules.")

		app.data.Lock()
		defer app.data.Unlock()

		if app.data.modules == nil {
			// concurrent close call
			return
		}
		// disable all modules before canceling context
		// if we will not do this any logic that takes our context can end
		// earlier so it won't disable modules properly
		for _, m := range app.data.modules.AllFromFront() {
			m.Disable()
			app.conf.Logger.Debug("disabled module.", "module", m.Name())
		}
	}()
	// after we disabled all modules we can close the context safely
	app.cancel()
	app.tr.Close()
	// save the user config
	app.saveUserConfig()
	// wait for additional things to end
	app.conf.Logger.Debug("waiting for waitgroup end...")
	app.wg.Wait()
	app.conf.Logger.Debug("waitgroup ended.")
	app.conf.Logger.Debug("closing window...")
	// close window last of all !!!
	if !main {
		//fixme:hack
		fyne.Do(app.closeWin)
		return nil
	}
	app.closeWin()
	return nil
}

func (app *App) closeWin() {
	app.data.Lock()
	defer app.data.Unlock()

	if app.data.win != nil {
		app.data.win.Close()
		app.conf.Logger.Debug("closed window.")
		return
	}
	// make sure app.win is never set to nil after initialized
	app.conf.Logger.Error("tried to close window before it was initialized (nil)")
}

//fixme very rarely the app process can freeze forever (while the fyne window closes)
//upd: it starts closing window but not ends, so the problem is probably fyne.DoAndWait
