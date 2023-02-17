package main

import (
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

type ProviderNode struct {
	input *ProviderNode
	sig   *types.Signature // current node signature
}
