package app

import (
	"github.com/elliotchance/orderedmap/v3"
	"github.com/something-that-is-cool/zutil/app/module"
)

func (app *App) createModulesFromConfigs(configs []module.Config) *modulesMap {
	type toCreateModule struct {
		Config   module.Config
		Property module.Property
	}
	modules := orderedmap.NewOrderedMap[module.Config, module.Module]()
	toCreate := make([]toCreateModule, 0, len(configs))

	func() {
		app.userConf.Lock()
		defer app.userConf.Unlock()
		// uConf must not be nil here because it is init func
		for _, conf := range configs {
			property := app.mustExtractPropertyUnsafe(conf)
			toCreate = append(toCreate, toCreateModule{Config: conf, Property: property})
		}
	}()
	// creating all modules within of the uConfMu !!!
	for _, m := range toCreate {
		modules.Set(m.Config, m.Config.Create(m.Property))
	}
	return modules
}

func (app *App) mustExtractPropertyUnsafe(conf module.Config) module.Property {
	property := conf.DefaultProperty()
	// check for value in user config
	v, ok := app.userConf.V.Modules[conf.Identifier()]
	if ok {
		// extract property from module value
		property = v
	} else {
		// add default property to config
		app.userConf.V.Modules[conf.Identifier()] = property
	}
	return property
}
