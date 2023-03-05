package comm

import (
	"go/types"
)

type BeanKind int

const (
	BuildInKind BeanKind = iota
	StructKind
	InterfaceKind
	PointerKind
)

type Bean struct {
	pkg string
	typ types.Type
}

func (b *Bean) PkgPath() string {
	return b.pkg
}

func (b *Bean) String() string {
	// TODO what to do for pointer kind?
	return b.typ.String()
}

func (b *Bean) Kind() BeanKind {
	return kindOf(b.typ)
}

func kindOf(typ types.Type) BeanKind {
	switch t := typ.(type) {
	case *types.Named:
		return kindOf(t.Underlying())
	case *types.Interface:
		return InterfaceKind
	case *types.Pointer:
		return PointerKind
	case *types.Struct:
		return StructKind
	default:
		return BuildInKind
	}
}

func (b *Bean) Identical(other *Bean) bool {
	if types.Identical(b.typ, other.typ) {
		return true
	}
	if b.String() == other.String() {
		return true
	}
	return false
}
