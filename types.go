package main

import (
	"go/ast"
	"go/types"

	"github.com/dave/dst/decorator"
)

type BeanID string

func NewBeanID(t types.Type) BeanID {
	return BeanID(t.String())
}

type FuncId string

func NewFuncID(sig *types.Func) FuncId {
	return FuncId(sig.String())
}

type PkgPath string

func NewPkgPath(pkg *decorator.Package) PkgPath {
	return PkgPath(pkg.PkgPath)
}

type LoadMode int

const (
	LoadProvder LoadMode = 1 << iota
	LoadInjector
)

type LoadConfig struct {
	LoadMode LoadMode
}

func (cfg *LoadConfig) IsInMode(mode LoadMode) bool {
	return cfg.LoadMode&mode == mode
}

type InjectorFunc struct {
	Pkg                 *decorator.Package // the package of injector
	Func                *types.Func        // injector func
	ProviderFuncs       []*types.Func      // provider list
	CallExpr            *ast.CallExpr      // the build call in injector body
	OnlyHasFuncProvider bool
}

func (injector *InjectorFunc) Build(di *DIContext) {
	for _, e := range injector.CallExpr.Args {
		obj := injector.Pkg.TypesInfo.ObjectOf(unwrap(e))
		if f, ok := obj.(*types.Func); ok {
			di.providers[NewFuncID(f)] = f
			injector.ProviderFuncs = append(injector.ProviderFuncs, f)
		} else {
			injector.OnlyHasFuncProvider = false
		}
	}
}

func (injector *InjectorFunc) PkgPathAt(i int) PkgPath {
	return PkgPath("TODO")
}

func (i *InjectorFunc) AddProvider(fn *types.Func) {
	log.Debug("add provider", "func", fn)
}
