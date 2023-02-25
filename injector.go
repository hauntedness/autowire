package main

import (
	"go/ast"
	"go/types"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/exp/slog"
)

// InjectorFunc provide convenient api to parse injector
type InjectorFunc struct {
	File                *dst.File
	Pkg                 *decorator.Package // the package of injector
	Func                *types.Func        // injector func
	ProviderFuncs       []*types.Func      // provider list
	CallExpr            *ast.CallExpr      // the call expression in wire.Build()
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
	slog.Debug("add provider", "func", fn)
}

func (i *InjectorFunc) Refactor() {
	for range i.File.Imports {
		// add more packages
	}
	for range i.CallExpr.Args {
		// add more args
	}
	// resolve conflict name
	i.Pkg.Save()
}
