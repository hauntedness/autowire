package main

import (
	"errors"
	"go/ast"
	"go/types"
	"os"
	"strings"
	"time"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/conf"
	"golang.org/x/exp/slog"
)

var (
	errorType   = types.Universe.Lookup("error").Type()
	cleanupType = types.NewSignature(nil, nil, nil, false)
)

func init() {
	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				a.Value = slog.StringValue(a.Value.Time().Format(time.DateTime))
			}
			return a
		},
	}
	log := slog.New(opts.NewTextHandler(os.Stdout))
	slog.SetDefault(log)
}

func main() {
	di := NewDIContext()
	cfg := &LoadConfig{LoadMode: LoadInjector | LoadProvder}
	err := di.LoadPackage("github.com/huantedness/autowire/example", cfg)
	if err != nil {
		panic(err)
	}
	// // check point 1
	// func() {
	// 	for _, v := range di.providers {
	// 		slog.Debug("found provider", "value", v)
	// 	}
	// 	for _, v := range di.injectors {
	// 		slog.Debug("found injector", "value", v)
	// 	}
	// }()
	di.RootSearch()
}

// DIContext entry to build injector quickly,
// automatically find all necessary providers and then rewrite source with full provider set
type DIContext struct {
	providers map[FuncId]*types.Func // only store function provider
	injectors map[FuncId]*InjectorFunc
	packages  map[PkgPath]*decorator.Package
	objects   map[ObjectId]types.Object
	funcs     map[FuncId]*types.Func
}

func NewDIContext() *DIContext {
	return &DIContext{
		providers: map[FuncId]*types.Func{},
		injectors: map[FuncId]*InjectorFunc{},
		packages:  map[PkgPath]*decorator.Package{},
		objects:   make(map[ObjectId]types.Object, 100),
		funcs:     make(map[FuncId]*types.Func, 100),
	}
}

func (di *DIContext) loadPackage(path PkgPath) (pkg *decorator.Package) {
	pkg = di.packages[path]
	if pkg != nil {
		return
	}
	pkgs, err := decorator.Load(conf.DefaultConf, path)
	if err != nil {
		panic(err)
	}
	if len(pkgs) == 0 {
		panic("can not load package " + path)
	}
	di.packages[path] = pkgs[0]
	for _, obj := range di.packages[path].TypesInfo.Defs {
		if obj != nil && obj.Pkg() != nil {
			if named, ok := obj.Type().(*types.Named); ok {
				// if the obj from other package, get the source pkg path
				namedObj := named.Obj()
				name := namedObj.Name()
				if namedObj.Pkg() == nil {
					// this happens on build in types such as error
					continue
				}
				id := NewObjectIdFromPkg(namedObj.Pkg(), name)
				// while currently we use obj value from current package
				di.objects[id] = namedObj
			} else {
				id := ObjectId{pkg: obj.Pkg().Path(), str: obj.Type().String()}
				di.objects[id] = obj
			}
		}
	}
	return pkgs[0]
}

// WalkLocalPackage load local injector and provider
func (di *DIContext) LoadPackage(pkgPath PkgPath, config *LoadConfig) error {
	pkg := di.loadPackage(pkgPath)
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
	converted := DstToAst[*dst.FuncDecl, *ast.FuncDecl](fd, pkg)
	// wire.Build(NewEvent, msg.NewMessage)
	callExpr, err := findInjectorBuild(pkg.TypesInfo, converted)
	if err != nil {
		return err
	}
	if callExpr == nil {
		slog.Info("not a builder", "fd", fd.Name)
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
}

func (di *DIContext) addInjector(v *InjectorFunc) {
	di.injectors[NewFuncID(v.Func)] = v
}

func (di *DIContext) addProvider(fn *types.Func) {
	di.providers[NewFuncID(fn)] = fn
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
	type BeanID = string
	providerFuncs := make(map[BeanID]*types.Signature)
	for _, f := range injectorFunc.ProviderFuncs {
		s := f.Type().(*types.Signature)
		out := s.Results().At(0).Type()
		providerFuncs[BeanID(out.String())] = s
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

func (di *DIContext) Get(id ObjectId) types.Object {
	if di.packages[id.pkg] == nil {
		di.LoadPackage(id.pkg, &LoadConfig{LoadMode: LoadProvder})
	}
	return di.objects[id]
}

func (di *DIContext) build(injector *InjectorFunc) {
	// eg: NewEvent(*msg.Message)
	funcSet := make(map[FuncId]*types.Func)
	for _, f := range injector.ProviderFuncs {
		// f something like NewEvent(*msg.Message)
		params := f.Type().(*types.Signature).Params()
		// check and add missing provider to funcSet
		for i := 0; i < params.Len(); i++ {
			// v something like *msg.Message or msg.SomeInterface
			v := params.At(i)
			id := FetchObjectId(v)
			obj := di.Get(id)
			if obj == nil {
				slog.Debug(v.String(), v)
			}
			provider := di.FindProvider(obj)
			funcSet[NewFuncID(provider)] = provider
		}
	}
	for k := range funcSet {
		slog.Info("find", "func id", k)
	}
}

func (di *DIContext) FindProvider(out types.Object) *types.Func {
	for _, p := range di.providers {
		v := p.Type().(*types.Signature).Results().At(0)
		id := FetchObjectId(v)
		obj := di.Get(id)
		if types.Identical(out.Type(), obj.Type()) {
			return p
		}
	}
	slog.Error("FindProvider", errors.New("cound not find provider"), "out", out)
	os.Exit(0)
	return nil
}
