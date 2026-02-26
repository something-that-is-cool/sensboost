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
	}
}

func (app *App) createControllerSensitivityModule(proc *win.Process) module.Config {
	conf := &modules.ControllerSensitivity{
		Process: proc,
		Error:   app.onError("controller_sensitivity"),
	}
	conf.OnChange = onModuleChange[float64](app, conf.Identifier())
	return conf
}

func (app *App) createNoDynamicFovModule(proc *win.Process) module.Config {
	conf := &modules.NoDynamicFov{
		Process: proc,
		Error:   app.onError("no_dynamic_fov"),
	}
	conf.OnChange = onModuleChange[bool](app, conf.Identifier())
	return conf
}

func (app *App) createNoHurtCamModule(proc *win.Process) module.Config {
	conf := &modules.NoHurtCam{
		Process: proc,
		Error:   app.onError("no_hurt_cam"),
	}
	conf.OnChange = onModuleChange[bool](app, conf.Identifier())
	return conf
}

func (app *App) createAutoSprintModule(proc *win.Process) module.Config {
	conf := &modules.AutoSprint{
		Process: proc,
		Error:   app.onError("auto_sprint"),
	}
	conf.OnChange = onModuleChange[bool](app, conf.Identifier())
	return conf
}

func (app *App) createNoParticleModule(proc *win.Process) module.Config {
	conf := &modules.NoParticle{
		Process: proc,
		Error:   app.onError("no_particle"),
	}
	conf.OnChange = onModuleChange[bool](app, conf.Identifier())
	return conf
}

func (app *App) onError(mod string) func(err error) {
	return func(err error) {
		app.conf.Logger.Error("an error occurred", "module", mod, "err", err.Error())
	}
}
