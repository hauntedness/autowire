package main

import (
	"fmt"
	"go/ast"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func unwrap(expr ast.Expr) *ast.Ident {
	switch v := expr.(type) {
	case *ast.Ident:
		return v
	case *ast.SelectorExpr:
		return unwrap(v.Sel)
	case *ast.StarExpr:
		return unwrap(v.X)
	default:
		panic("no match" + fmt.Sprintf("%v", v))
	}
}

func convertAD[L ast.Node, R dst.Node](left L, pkg *decorator.Package) R {
	n := pkg.Decorator.Dst.Nodes[left]
	return n.(R)
}

func convertDA[L dst.Node, R ast.Node](left L, pkg *decorator.Package) R {
	n := pkg.Decorator.Ast.Nodes[left]
	return n.(R)
}
