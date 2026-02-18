package app

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var controllinURL = func() *url.URL {
	u, err := url.Parse("https://t.me/+rweTeGr1vOxjM2Qy")
	if err != nil {
		panic(fmt.Errorf("parse controllin link: %w", err))
	}
	return u
}()

func (app *App) createContent(proc *win.Process) (fyne.CanvasObject, []module.Module, error) {
	m := []module.Module{
		app.createControllerSensitivityModule(proc),
		app.createNoDynamicFovModule(proc),
		app.createNoHurtCamModule(proc),
		app.createAutoSprintModule(proc),
		app.createNoParticleModule(proc),
	}
	var obj []fyne.CanvasObject
	for _, m := range m {
		box := container.NewVBox(m.CreateObjects()...)
		window := container.NewInnerWindow(m.Name(), box)

		obj = append(obj, window)
	}
	b := container.NewBorder(
		container.NewGridWithRows(4, obj...),
		app.createFooter(),
		nil, nil,
	)
	return b, m, nil
}

func (app *App) createFooter() fyne.CanvasObject {
	return container.NewBorder(
		nil, nil,
		widget.NewHyperlink("Join to controllin", controllinURL),
		widget.NewLabel("Ivan Zov 2011"),
	)
}

func (app *App) createControllerSensitivityModule(proc *win.Process) module.Module {
	return (&modules.ControllerSensitivity{
		Process: proc,
		Error:   app.onError("controller_sensitivity"),
	}).Create()
}

func (app *App) createNoDynamicFovModule(proc *win.Process) module.Module {
	return (&modules.NoDynamicFov{
		Process: proc,
		Error:   app.onError("no_dynamic_fov"),
	}).Create()
}

func (app *App) createNoHurtCamModule(proc *win.Process) module.Module {
	return (&modules.NoHurtCam{
		Process: proc,
		Error:   app.onError("no_hurt_cam"),
	}).Create()
}

func (app *App) createAutoSprintModule(proc *win.Process) module.Module {
	return (&modules.AutoSprint{
		Process: proc,
		Error:   app.onError("auto_sprint"),
	}).Create()
}

func (app *App) createNoParticleModule(proc *win.Process) module.Module {
	return (&modules.NoParticle{
		Process: proc,
		Error:   app.onError("no_particle"),
	}).Create()
}

func (app *App) onError(mod string) func(err error) {
	return func(err error) {
		app.conf.Logger.Error("an error occurred", "module", mod, "err", err.Error())
	}
}
