package main

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/exp/slog"
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

func DstToAst[L dst.Node, R ast.Node](left L, pkg *decorator.Package) R {
	n := pkg.Decorator.Ast.Nodes[left]
	return n.(R)
}

// funcOf convert fd to *types.Func
func funcOf(fd *dst.FuncDecl, pkg *decorator.Package) *types.Func {
	reverted := DstToAst[*dst.FuncDecl, *ast.FuncDecl](fd, pkg)
	_type := pkg.TypesInfo.ObjectOf(reverted.Name)
	return _type.(*types.Func)
}

// fetch package, object from function args or results
func FetchObjectId(v *types.Var) ObjectId {
	origin := v.Origin().Type()
	switch typ := origin.(type) {
	case *types.Named:
		// named interface
		obj := typ.Obj()
		id := NewObjectIdFromPkg(obj.Pkg(), obj.Name())
		return id
	case *types.Interface:
		id := ObjectId{pkg: v.Pkg().Path(), str: typ.String()}
		return id
	case *types.Pointer:
		id := ObjectId{pkg: v.Pkg().Path(), str: typ.String()}
		return id
	case *types.Struct:
		id := ObjectId{pkg: v.Pkg().Path(), str: typ.String()}
		return id
	default:
		// scenarios which is not covered
		slog.Info("something impossible", typ)
		panic(typ)
	}
}
