package pkg

import (
	"go/types"
	"strings"

	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/logs"
	"github.com/huantedness/autowire/pkg/comm"
)

type ProcessConfig struct {
	RewriteSource     bool
	ProviderPredicate func(fn *types.Func) bool // used to report whether a function can be provider
}

var defaultProcessConfig = &ProcessConfig{
	RewriteSource: false,
	ProviderPredicate: func(fn *types.Func) bool {
		return strings.HasPrefix(fn.Name(), "New")
	},
}

type DIContext struct {
	conf      *ProcessConfig
	pkgs      map[string]*decorator.Package
	files     map[objRef]*comm.WireFile
	injectors map[objRef]*comm.Injector
	providers map[objRef]*comm.Provider // maybe here better be a map[BeanId]map[FuncId]*Provider
}

func NewDIContext(conf *ProcessConfig) *DIContext {
	if conf == nil {
		conf = defaultProcessConfig
	}
	return &DIContext{
		conf:      conf,
		pkgs:      map[string]*decorator.Package{},
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
		for i := range [100]struct{}{} {
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
					logs.Debug("compare", "b", b.String(), "bean", bean.String())
					if b.Identical(bean) {
						inj.AddProvider(p)
					}
				}
			}
			if i == 100 {
				logs.Warn("fail to find provider after tried 1000 times", "injector", inj)
			}
		}
	}
}

func (di *DIContext) refactor() {
	for _, file := range di.files {
		file.Refactor()
	}
	for path, pkg := range di.pkgs {
		if di.conf.RewriteSource {
			logs.Debug("saving", "package", path)
			err := pkg.Save()
			if err != nil {
				panic(err)
			}
		}
	}
}
