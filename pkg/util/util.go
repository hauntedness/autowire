package util

import (
	"fmt"
	"go/ast"
)

func Unwrap(expr ast.Expr) *ast.Ident {
	switch v := expr.(type) {
	case *ast.Ident:
		return v
	case *ast.SelectorExpr:
		return Unwrap(v.Sel)
	case *ast.StarExpr:
		return Unwrap(v.X)
	default:
		panic(fmt.Errorf("error get ident from expr %v", expr))
	}
}
