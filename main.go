package main

import (
	"go/ast"
	"go/types"
	"os"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/exp/slog"
	"golang.org/x/tools/go/packages"
)

var (
	log         *slog.Logger
	errorType   = types.Universe.Lookup("error").Type()
	cleanupType = types.NewSignature(nil, nil, nil, false)
	conf        = &packages.Config{
		BuildFlags: []string{"-tags=wireinject"},
		Mode:       packages.NeedName | packages.NeedFiles | packages.NeedDeps | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}
)

func init() {
	opts := slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}
	log = slog.New(opts.NewTextHandler(os.Stdout))
}

func main() {
	di := NewDIContext()
	cfg := &LoadConfig{LoadMode: LoadInjector | LoadProvder}
	err := di.LoadPackage("github.com/huantedness/autowire/example", cfg)
	if err != nil {
		panic(err)
	}
	// check point 1
	func() {
		for _, t := range di.beans {
			log.Debug("found bean", "type", t)
		}
		for _, v := range di.cache {
			log.Debug("found cache", "value", v)
		}
		for _, v := range di.providers {
			log.Debug("found provider", "value", v)
		}
		for _, v := range di.injectors {
			log.Debug("found injector", "value", v)
		}
	}()
	di.RootSearch()
}

// DIContext entry to build injector quickly,
// automatically find all necessary providers and then rewrite source with full provider set
type DIContext struct {
	beans     map[BeanID]types.Type
	providers map[FuncId]*types.Func // only store function provider
	injectors map[FuncId]*InjectorFunc
	cache     map[BeanID]map[FuncId]*types.Func // to find provider and injector quickly
	packages  map[PkgPath]*decorator.Package
}

func NewDIContext() *DIContext {
	return &DIContext{
		beans:     map[BeanID]types.Type{},
		providers: map[FuncId]*types.Func{},
		injectors: map[FuncId]*InjectorFunc{},
		cache:     map[BeanID]map[FuncId]*types.Func{},
		packages:  map[PkgPath]*decorator.Package{},
	}
}

func (di *DIContext) loadPackage(path PkgPath) (pkg *decorator.Package) {
	pkg = di.packages[path]
	if pkg != nil {
		return
	}
	pkgs, err := decorator.Load(conf, string(path))
	if err != nil {
		panic(err)
	}
	if len(pkgs) == 0 {
		panic("can not load package " + path)
	}
	di.packages[path] = pkgs[0]
	return pkgs[0]
}

// WalkLocalPackage load local injector and provider
func (di *DIContext) LoadPackage(pkgPath PkgPath, config *LoadConfig) error {
	pkg := di.loadPackage(PkgPath(pkgPath))
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*dst.FuncDecl); ok {
				if config.IsInMode(LoadProvder) {
					di.LoadProvider(fd, pkg)
				}
				if config.IsInMode(LoadInjector) {
					err := di.LoadInjector(fd, pkg)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// LoadInjector load fd as injector if it passes the validation
func (di *DIContext) LoadInjector(fd *dst.FuncDecl, pkg *decorator.Package) error {
	converted := revertToAstNode[*dst.FuncDecl, *ast.FuncDecl](fd, pkg)
	// wire.Build(NewEvent, msg.NewMessage)
	callExpr, err := findInjectorBuild(pkg.TypesInfo, converted)
	if err != nil {
		return err
	}
	if callExpr == nil {
		log.Info("not a builder", "fd", fd.Name)
		return nil
	}
	fn := funcOf(fd, pkg)
	injector := &InjectorFunc{
		Pkg:                 pkg,
		Func:                fn,
		ProviderFuncs:       []*types.Func{},
		CallExpr:            callExpr,
		OnlyHasFuncProvider: true,
	}
	injector.Build(di)
	di.addInjector(injector)
	return nil
}

// LoadProvider load fd as provider if it passes the validation
func (di *DIContext) LoadProvider(fd *dst.FuncDecl, pkg *decorator.Package) {
	funcName := fd.Name.Name
	if !strings.HasPrefix(funcName, "New") {
		return
	}
	fn := funcOf(fd, pkg)
	sig := fn.Type().(*types.Signature)
	outputSignature, err := funcOutput(sig)
	if err != nil {
		slog.Debug("function is not a potential provider", "outputSignature", outputSignature)
		return
	}
	di.addProvider(fn)
	bean := outputSignature.out
	di.addBean(bean)
	di.addCache(bean, outputSignature, fn)
}

func (di *DIContext) addInjector(v *InjectorFunc) {
	di.injectors[NewFuncID(v.Func)] = v
}

func (di *DIContext) addProvider(fn *types.Func) {
	di.providers[NewFuncID(fn)] = fn
}

func (di *DIContext) addCache(bean types.Type, outputSignature outputSignature, fn *types.Func) {
	cached := di.cache[NewBeanID(bean)]
	if cached == nil {
		cached = make(map[FuncId]*types.Func)
		cached[NewFuncID(fn)] = fn
	}
	di.cache[NewBeanID(bean)] = cached
}

func (di *DIContext) addBean(bean types.Type) {
	di.beans[NewBeanID(bean)] = bean
}

// RootSearch find and make up incomplete injectors
func (di *DIContext) RootSearch() {
	for _, injector := range di.injectors {
		complete := di.isWireReady(injector)
		if !complete {
			di.build(injector)
		}
	}
}

// check injector builder, if some
func (di *DIContext) isWireReady(injectorFunc *InjectorFunc) bool {
	// only handle function provider, for other cases, will use native wire
	if !injectorFunc.OnlyHasFuncProvider {
		return true
	}
	// packagePath := fn.Pkg().Path()
	// pkg := di.loadPackage(PkgPath(packagePath))
	sig := injectorFunc.Func.Type().(*types.Signature)
	providerFuncs := make(map[BeanID]*types.Signature)
	for _, f := range injectorFunc.ProviderFuncs {
		s := f.Type().(*types.Signature)
		v := s.Results().At(0).Type()
		providerFuncs[BeanID(v.String())] = s
	}
	res := sig.Results().At(0).Type().String()
	rootPath := BeanID(res)
	_, ok := providerFuncs[rootPath]
	if !ok {
		panic("need at least one provider with return type " + rootPath)
	}
	complete := true
	for _, f := range injectorFunc.ProviderFuncs {
		params := f.Type().(*types.Signature).Params()
		for i := 0; i < params.Len(); i++ {
			pi := params.At(i)
			if _, ok := providerFuncs[BeanID(pi.Type().String())]; !ok {
				return false
			}
		}
	}
	return complete
}

func (di *DIContext) build(injector *InjectorFunc) {
	// eg: NewEvent(*msg.Message)
	// for each NewEvent provider
	// find a message provider and refactor the code if message not present
	for _, f := range injector.ProviderFuncs {
		params := f.Type().(*types.Signature).Params()
		for i := 0; i < params.Len(); i++ {
			v := params.At(i)
			path := injector.PkgPathAt(i)
			conf := &LoadConfig{
				LoadMode: LoadProvder,
			}
			err := di.LoadPackage(PkgPath(path), conf)
			if err != nil {
				panic(err)
			}
			fn := di.Pick(v.Type())
			injector.AddProvider(fn)
		}
	}
}

func (di *DIContext) Pick(typ types.Type) (provider *types.Func) {
	beans := di.cache[NewBeanID(typ)]
	for id, bean := range beans {
		// local first
		if strings.Contains(string(id), "") {
			return bean
		}
		provider = bean
	}
	return
}
