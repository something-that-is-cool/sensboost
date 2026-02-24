package app

import (
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

func (app *App) setupModules(proc *win.Process) []module.Module {
	return []module.Module{
		app.createControllerSensitivityModule(proc),
		app.createNoDynamicFovModule(proc),
		app.createNoHurtCamModule(proc),
		app.createAutoSprintModule(proc),
		app.createNoParticleModule(proc),
	}
}

func (app *App) createControllerSensitivityModule(proc *win.Process) module.Module {
	return modules.ControllerSensitivity{
		Process: proc,
		Error:   app.onError("controller_sensitivity"),
	}.Create()
}

func (app *App) createNoDynamicFovModule(proc *win.Process) module.Module {
	return modules.NoDynamicFov{
		Process: proc,
		Error:   app.onError("no_dynamic_fov"),
	}.Create()
}

func (app *App) createNoHurtCamModule(proc *win.Process) module.Module {
	return modules.NoHurtCam{
		Process: proc,
		Error:   app.onError("no_hurt_cam"),
	}.Create()
}

func (app *App) createAutoSprintModule(proc *win.Process) module.Module {
	return modules.AutoSprint{
		Process: proc,
		Error:   app.onError("auto_sprint"),
	}.Create()
}

func (app *App) createNoParticleModule(proc *win.Process) module.Module {
	return modules.NoParticle{
		Process: proc,
		Error:   app.onError("no_particle"),
	}.Create()
}
