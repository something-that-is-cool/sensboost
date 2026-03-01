package app

import (
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

func (app *App) setupModules(proc *win.Process) []module.Config {
	return []module.Config{
		app.createControllerSensitivityModule(proc),
		app.createNoDynamicFovModule(proc),
		app.createNoHurtCamModule(proc),
		app.createAutoSprintModule(proc),
		app.createNoParticleModule(proc),
		app.createNoFireModule(proc),
		app.createZoomModule(proc),
	}
}

func (app *App) createControllerSensitivityModule(proc *win.Process) module.Config {
	conf := &modules.ControllerSensitivity{
		Process: proc,
		Error:   app.onError("controller_sensitivity"),
	}
	conf.OnValueChanged = onModuleValueChanged[float64](app, conf.Identifier())
	return conf
}

func (app *App) createNoDynamicFovModule(proc *win.Process) module.Config {
	conf := &modules.NoDynamicFov{
		Process: proc,
		Error:   app.onError("no_dynamic_fov"),
	}
	conf.OnToggle = app.onModuleToggled(conf.Identifier())
	return conf
}

func (app *App) createNoHurtCamModule(proc *win.Process) module.Config {
	conf := &modules.NoHurtCam{
		Process: proc,
		Error:   app.onError("no_hurt_cam"),
	}
	conf.OnToggle = app.onModuleToggled(conf.Identifier())
	return conf
}

func (app *App) createAutoSprintModule(proc *win.Process) module.Config {
	conf := &modules.AutoSprint{
		Process: proc,
		Error:   app.onError("auto_sprint"),
	}
	conf.OnToggle = app.onModuleToggled(conf.Identifier())
	return conf
}

func (app *App) createNoParticleModule(proc *win.Process) module.Config {
	conf := &modules.NoParticle{
		Process: proc,
		Error:   app.onError("no_particle"),
	}
	conf.OnToggle = app.onModuleToggled(conf.Identifier())
	return conf
}

func (app *App) createNoFireModule(proc *win.Process) module.Config {
	conf := &modules.NoFire{
		Process: proc,
		Error:   app.onError("no_fire"),
	}
	conf.OnToggle = app.onModuleToggled(conf.Identifier())
	return conf
}

func (app *App) createZoomModule(proc *win.Process) module.Config {
	conf := &modules.Zoom{
		Process: proc,
		Error:   app.onError("zoom"),
	}
	conf.OnToggle = app.onModuleToggled(conf.Identifier())
	return conf
}

func (app *App) onError(mod string) func(err error) {
	return func(err error) {
		app.conf.Logger.Error("an error occurred", "module", mod, "err", err.Error())
	}
}
