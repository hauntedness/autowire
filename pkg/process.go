package pkg

import (
	"go/types"

	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/logs"
	"github.com/huantedness/autowire/pkg/comm"
)

// optional config when processing autowire
type ProcessConfigurer interface {
	//
	// whether to rewrite source code
	WillRewriteSource() bool

	// used to report whether a function is an injector
	// this is helpful when you only want to generate a part of functions in the package
	InjectorPredicate(fn *types.Func) bool

	// used to report whether a function can be provider
	ProviderPredicate(fn *types.Func) bool

	// used to find proper provider from multiple ones
	ProviderElect(inj *comm.Injector, bean *comm.Bean, providers map[string]*comm.Provider) *comm.Provider
}

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
		for range [1000]struct{}{} {
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
	for _, file := range di.files {
		file.Refactor()
	}
	for path, pkg := range di.pkgs {
		if di.conf.WillRewriteSource() {
			logs.Debug("saving package", "package", path)
			err := pkg.Save()
			if err != nil {
				panic(err)
			}
		}
	}
}
