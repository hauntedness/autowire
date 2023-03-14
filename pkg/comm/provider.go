package comm

import (
	"go/types"

	"github.com/huantedness/autowire/logs"
)

type Provider struct {
	fn *types.Func
}

func NewProvider(fn *types.Func) *Provider {
	return &Provider{fn: fn}
}

func (p *Provider) Require() []*Bean {
	type BeanId = string
	ret := make([]*Bean, 0, 3)
	params := p.fn.Type().(*types.Signature).Params()
	for i := range make([]struct{}, params.Len()) {
		v := params.At(i)
		bean := p.fromVar(v)
		ret = append(ret, bean)
	}
	return ret
}

func (p *Provider) Provide() *Bean {
	result := p.fn.Type().(*types.Signature).Results()
	bean := p.fromVar(result.At(0))
	return bean
}

func (p *Provider) Name() string {
	return p.fn.Name()
}

func (p *Provider) String() string {
	return p.fn.String()
}

func (p *Provider) Package() string {
	return p.fn.Pkg().Path()
}

// getBean convert param or result to Bean
func (p *Provider) fromVar(v *types.Var) *Bean {
	origin := v.Origin().Type()
	switch typ := origin.(type) {
	case *types.Named:
		// named interface or struct or pointer
		obj := typ.Obj()
		pkg := obj.Pkg()
		path := ""
		if pkg != nil {
			path = pkg.Path()
		}
		bean := &Bean{pkg: path, typ: typ}
		return bean
	case *types.Interface:
		// same as caller package
		bean := &Bean{pkg: p.fn.Pkg().Path(), typ: typ}
		return bean
	case *types.Pointer:
		bean := &Bean{pkg: deref(p.fn.Pkg().Path(), typ), typ: typ}
		return bean
	case *types.Struct:
		bean := &Bean{pkg: p.fn.Pkg().Path(), typ: typ}
		return bean
	default:
		logs.Warn("something should not happen!", "type", typ)
		panic(typ)
	}
}

func deref(currentPkg string, typ types.Type) string {
	switch t := typ.(type) {
	case *types.Pointer:
		return deref(currentPkg, t.Elem())
	case *types.Named:
		pkg := t.Obj().Pkg()
		path := ""
		if pkg != nil {
			path = pkg.Path()
		}
		return path
	default:
		return currentPkg
	}
}
