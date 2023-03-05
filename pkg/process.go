package pkg

import (
	"go/types"

	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/pkg/comm"
)

type DIContext struct {
	pkgs      map[string]*decorator.Package
	objects   map[objRef]types.Object
	injectors map[objRef]*comm.Injector
	providers map[objRef]*comm.Provider
}

func NewDIContext() *DIContext {
	return &DIContext{
		pkgs:      map[string]*decorator.Package{},
		objects:   map[objRef]types.Object{},
		injectors: map[objRef]*comm.Injector{},
		providers: map[objRef]*comm.Provider{},
	}
}

func (di *DIContext) Process(path string) {
	pkg := loadPackage(path)
	// parse
	config := &LoadConfig{
		LoadMode: LoadProvider | LoadInjector,
	}
	// load entry package
	di.loadProviderAndInjector(pkg, config)

	di.doInject()
}

// doInject process each injector,
func (di *DIContext) doInject() {
	for _, inj := range di.injectors {
		// here mean all required is provided
		for done := false; !done; {
			m := inj.Require()
			if len(m) == 0 {
				done = true
				break
			}
			for _, bean := range m {
				path := bean.PkgPath()
				if di.pkgs[path] == nil {
					pkg := loadPackage(path)
					di.pkgs[path] = pkg
					di.loadProviderAndInjector(pkg, &LoadConfig{LoadMode: LoadProvider})
				}
				// TODO here to find a bean provider
				for _, p := range di.providers {
					b := p.Provide()
					if b.Identical(bean) {
						inj.AddProvider(p)
					}
				}
				done = false
			}
		}
	}
}
