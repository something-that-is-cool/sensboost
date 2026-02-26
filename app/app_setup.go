package app

import "github.com/something-that-is-cool/zutil/app/module"

func (app *App) createModulesFromConfigs(configs []module.Config) []module.Module {
	type toCreateModule struct {
		Config   module.Config
		Property module.Property
	}
	modules := make([]module.Module, 0, len(configs))
	toCreate := make([]toCreateModule, 0, len(configs))

	func() {
		app.uConf.RLock()
		defer app.uConf.RUnlock()

		for _, conf := range configs {
			property := app.mustExtractPropertyUnsafe(conf)
			toCreate = append(toCreate, toCreateModule{Config: conf, Property: property})
		}
	}()
	// creating all modules within of the mutex !!!
	for _, m := range toCreate {
		modules = append(modules, m.Config.Create(m.Property))
	}
	return modules
}

func (app *App) mustExtractPropertyUnsafe(conf module.Config) module.Property {
	property := conf.DefaultProperty()
	// check for value in user config
	v, ok := app.uConf.Modules[conf.Identifier()]
	if ok {
		// extract property from module value
		property = propertyFromValue(v)
	} else {
		// add default property to config
		app.uConf.Modules[conf.Identifier()] = property
	}
	return property
}
