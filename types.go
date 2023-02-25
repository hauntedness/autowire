package main

import (
	"go/ast"
	"go/types"

	"github.com/dave/dst/decorator"
)

type ObjectId struct {
	pkg PkgPath
	str string
}

func (id ObjectId) String() string {
	return id.str
}

func NewObjectIdFromPkg(pkg *types.Package, name string) ObjectId {
	if pkg == nil {
		panic("should not happen")
	}
	return ObjectId{pkg: pkg.Path(), str: pkg.Path() + "." + name}
}

func NewObjectId(pkg PkgPath, ident *ast.Ident) ObjectId {
	return ObjectId{pkg: pkg, str: pkg + "." + ident.String()}
}

type FuncId = string

func NewFuncID(sig *types.Func) FuncId {
	return FuncId(sig.String())
}

type PkgPath = string

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
