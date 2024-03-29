package pkg

import (
	"log/slog"

	"github.com/dave/dst/decorator"
	"github.com/hauntedness/autowire/pkg/comm"
)

type DIContext struct {
	conf      ProcessConfigurer
	pkgs      map[string]*decorator.Package
	files     map[objRef]*comm.WireFile
	injectors map[objRef]*comm.Injector
	providers map[objRef]*comm.Provider // maybe here better be a map[BeanId]map[FuncId]*Provider
}

func NewDIContext(conf ProcessConfigurer) *DIContext {
	if conf == nil {
		conf = &DefaultProcessConfigurer{willRewriteSource: true}
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
	di.pkgs[pkg.PkgPath] = pkg
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
		// until all required is provided
		// while it is possible that some providers miss
		for i := range [1000]struct{}{} {
			m := inj.Require()
			if len(m) == 0 {
				break
			}
			if i == 999 {
				slog.Warn("still could not find privder after trying many times", "round", i)
				break
			}
			for _, bean := range m {
				path := bean.PkgPath()
				if di.pkgs[path] == nil {
					pkg := loadPackage(path)
					di.pkgs[path] = pkg
					di.loadProviderAndInjector(pkg, &LoadConfig{LoadMode: LoadProvider})
				}
				candidates := make(map[string]*comm.Provider)
				// TODO here we need more efficient way
				for _, p := range di.providers {
					b := p.Provide()
					if b.Identical(bean) {
						candidates[p.String()] = p
					}
				}
				p := di.conf.ProviderElect(inj, bean, candidates)
				inj.AddProvider(p)
			}
		}
	}
}

func (di *DIContext) refactor() {
	refactored := map[string]bool{}
	for _, file := range di.files {
		file.Refactor()
		refactored[file.Package()] = true
	}
	for path, pkg := range di.pkgs {
		if di.conf.WillRewriteSource() && refactored[path] {
			slog.Info("saving package", "package", path)
			err := pkg.Save()
			if err != nil {
				panic(err)
			}
		}
	}
}
