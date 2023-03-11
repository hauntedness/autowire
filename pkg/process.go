package pkg

import (
	"fmt"
	"go/types"

	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/pkg/comm"
	"golang.org/x/exp/slog"
)

type DIContext struct {
	pkgs      map[string]*decorator.Package
	objects   map[objRef]types.Object
	files     map[objRef]*comm.WireFile
	injectors map[objRef]*comm.Injector
	providers map[objRef]*comm.Provider // here better be a map[BeanId]map[FuncId]*Provider
}

func NewDIContext() *DIContext {
	return &DIContext{
		pkgs:      map[string]*decorator.Package{},
		objects:   map[objRef]types.Object{},
		files:     map[objRef]*comm.WireFile{},
		injectors: map[objRef]*comm.Injector{},
		providers: map[objRef]*comm.Provider{},
	}
}

func (di *DIContext) Process(path string) {
	pkg := loadPackage(path)
	di.pkgs[path] = pkg
	// parse
	config := &LoadConfig{
		LoadMode: LoadProvider | LoadInjector,
	}
	// load entry package
	di.loadProviderAndInjector(pkg, config)

	di.doInject()

	di.refactor()
}

// doInject process each injector,
func (di *DIContext) doInject() {
	for _, inj := range di.injectors {
		// here mean all required is provided
		for i := 0; i < 100; i++ {
			m := inj.Require()
			if len(m) == 0 {
				break
			}
			for _, bean := range m {
				path := bean.PkgPath()
				if di.pkgs[path] == nil {
					pkg := loadPackage(path)
					di.pkgs[path] = pkg
					di.loadProviderAndInjector(pkg, &LoadConfig{LoadMode: LoadProvider})
				}
				// TODO here to find a proper bean provider
				for _, p := range di.providers {
					b := p.Provide()
					slog.Info("compare", "b", b.String(), "bean", bean.String())
					if b.Identical(bean) {
						inj.AddProvider(p)
					}
				}
			}
			if i == 100 {
				slog.Warn("fail to find provider after tried 1000 times", "injector", inj)
			}
		}
	}
}

func (di *DIContext) refactor() {
	for ref, file := range di.files {
		fmt.Println(ref)
		file.Refactor()
	}
	for k, p := range di.pkgs {
		slog.Info("saving", "package", k)
		p.Save()
	}
}
