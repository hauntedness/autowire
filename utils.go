package main

import (
	"fmt"
	"go/ast"
	"go/types"

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
	case *ast.Field:
		return unwrap(v.Names[0])
	default:
		panic("no match" + fmt.Sprintf("%v", v))
	}
}

func revertToAstNode[L dst.Node, R ast.Node](left L, pkg *decorator.Package) R {
	n := pkg.Decorator.Ast.Nodes[left]
	return n.(R)
}

// funcOf convert fd to *types.Func
func funcOf(fd *dst.FuncDecl, pkg *decorator.Package) *types.Func {
	reverted := revertToAstNode[*dst.FuncDecl, *ast.FuncDecl](fd, pkg)
	_type := pkg.TypesInfo.ObjectOf(reverted.Name)
	return _type.(*types.Func)
}

func packageOf(typ types.Type) PkgPath {
	switch t := typ.(type) {
	case *types.Named:
		return PkgPath(t.Obj().Pkg().Path())
	case *types.Array, *types.Basic, *types.Chan, *types.Map, *types.Slice:
		return ""
	case *types.Pointer:
		return packageOf(t.Elem())
	case *types.Interface:
		return PkgPath(t.String())
	case *types.Struct:
		return PkgPath(t.String())
	default:
		panic("not supported type:" + t.String())
	}
}
