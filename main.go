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

var log *slog.Logger

func init() {
	opts := slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}
	log = slog.New(opts.NewTextHandler(os.Stdout))
}

var conf = &packages.Config{
	BuildFlags: []string{"-tags=wireinject"},
	Mode:       packages.NeedName | packages.NeedFiles | packages.NeedDeps | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
}

func main() {
	pkgs, err := decorator.Load(conf, "github.com/huantedness/autowire/example")
	if err != nil {
		panic(err)
	}
	di := NewDIContext()
	for _, pkg := range pkgs {
		di.packages[NewPkgPath(pkg)] = pkg
		err = di.WalkLocalPackage(pkg)
		if err != nil {
			panic(err)
		}
	}
	// check point 1
	func() {
		for _, t := range di.beans {
			log.Debug("found bean: ", "type", t)
		}
		for _, v := range di.cache {
			log.Debug("found cache: ", "value", v)
		}
		for _, v := range di.providers {
			log.Debug("found provider: ", "value", v)
		}
		for _, v := range di.injectors {
			log.Debug("found injector: ", "value", v)
		}
	}()
	// di.SearchAddtionPackages()
}

var (
	errorType   = types.Universe.Lookup("error").Type()
	cleanupType = types.NewSignature(nil, nil, nil, false)
)

// DIContext entry to build injector quickly,
// automatically find all necessary providers and then rewrite source with full provider set
type DIContext struct {
	beans     map[BeanID]types.Type
	providers map[FuncId]*types.Func // only store function provider
	injectors map[FuncId]*types.Func
	cache     map[BeanID]map[FuncId]*types.Func // to find provider and injector quickly
	packages  map[PkgPath]*decorator.Package
}

func NewDIContext() *DIContext {
	return &DIContext{
		beans:     map[BeanID]types.Type{},
		providers: map[FuncId]*types.Func{},
		injectors: map[FuncId]*types.Func{},
		cache:     map[BeanID]map[FuncId]*types.Func{},
		packages:  map[PkgPath]*decorator.Package{},
	}
}

func (di *DIContext) load(path PkgPath) {
	if di.packages[path] != nil {
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
}

// WalkLocalPackage load local injector and provider
func (di *DIContext) WalkLocalPackage(pkg *decorator.Package) error {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*dst.FuncDecl); ok {
				di.LoadProvider(fd, pkg)
				err := di.LoadInjector(fd, pkg)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// LoadInjector load fd as injector if it passes the validation
func (di *DIContext) LoadInjector(fd *dst.FuncDecl, pkg *decorator.Package) error {
	converted := convertDA[*dst.FuncDecl, *ast.FuncDecl](fd, pkg)
	// wire.Build(NewEvent, msg.NewMessage)
	callExpr, err := findInjectorBuild(pkg.TypesInfo, converted)
	if err != nil {
		return err
	}
	if callExpr == nil {
		log.Info("not a builder", "fd", fd.Name)
		return nil
	}
	fn := di.funcOf(fd, pkg)
	di.addInjector(fn)
	for _, e := range callExpr.Args {
		sig := pkg.TypesInfo.ObjectOf(unwrap(e))
		f := sig.(*types.Func)
		di.providers[NewFuncID(f)] = f
	}
	return nil
}

// funcOf convert fd to *types.Func
func (di *DIContext) funcOf(fd *dst.FuncDecl, pkg *decorator.Package) *types.Func {
	fd2 := convertDA[*dst.FuncDecl, *ast.FuncDecl](fd, pkg)
	_type := pkg.TypesInfo.ObjectOf(fd2.Name)
	return _type.(*types.Func)
}

// LoadProvider load fd as provider if it passes the validation
func (di *DIContext) LoadProvider(fd *dst.FuncDecl, pkg *decorator.Package) {
	funcName := fd.Name.Name
	if !strings.HasPrefix(funcName, "New") {
		return
	}
	fn := di.funcOf(fd, pkg)
	sig := fn.Type().(*types.Signature)
	outputSignature, err := funcOutput(sig)
	if err != nil {
		slog.Debug("function is not a potential provider", "outputSignature", outputSignature)
		return
	}
	for _, f := range fd.Type.Params.List {
		n := convertDA[*dst.Ident, *ast.Ident](f.Names[0], pkg)
		t := pkg.TypesInfo.ObjectOf(n)
		path := t.Pkg().Path()
		di.load(PkgPath(path))
		log.Debug(path)
	}
	di.addProvider(fn)
	bean := outputSignature.out
	di.addBean(bean)
	di.addCache(bean, outputSignature, fn)
}

func (di *DIContext) addInjector(sig *types.Func) {
	di.injectors[NewFuncID(sig)] = sig
}

func (di *DIContext) addProvider(sig *types.Func) {
	di.providers[NewFuncID(sig)] = sig
}

func (di *DIContext) addCache(bean types.Type, outputSignature outputSignature, fn *types.Func) {
	cached := di.cache[NewBeanID(bean)]
	if cached == nil {
		cached = make(map[FuncId]*types.Func)
		cached[NewFuncID(fn)] = fn
	}
}

func (di *DIContext) addBean(bean types.Type) {
	di.beans[NewBeanID(bean)] = bean
}

// SearchAddtionPackages do additional enrich after the 1st round as there might be some beans lack of providers
func (di *DIContext) SearchAddtionPackages() {
	panic("not implemented")
}
