package pkg

import (
	"go/types"
)

func matchObject(obj types.Object) {
	switch obj.(type) {
	case *types.Func:
	case *types.TypeName:
	case *types.PkgName, *types.Const, *types.Var, *types.Label, *types.Builtin, *types.Nil:
	default:
	}
}

func matchType(typ types.Type) {
	switch typ.(type) {
	case *types.Named:
	case *types.Interface:
	case *types.Signature:
	case *types.Array:
	case *types.Basic:
	case *types.Chan:
	case *types.Map:
	case *types.Pointer:
	case *types.Slice:
	case *types.Struct:
	case *types.Tuple:
	case *types.TypeParam:
	case *types.Union:
	default:

	}
}
