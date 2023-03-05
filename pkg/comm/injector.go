package comm

import (
	"go/types"
)

type FuncId = string

// Injector store
type Injector struct {
	fn        *types.Func
	origin    map[FuncId]*Provider // original providers from the injector source code
	providers map[FuncId]*Provider // all providers after analyzed
	auto      bool                 // whether autowire can fill up this provider
}

type BeanId = string

// NewInjector
func NewInjector(fn *types.Func, origin map[FuncId]*Provider, auto bool) *Injector {
	copy := make(map[FuncId]*Provider)
	for k, p := range origin {
		copy[k] = p
	}
	return &Injector{
		fn:        fn,
		origin:    origin,
		providers: copy,
		auto:      auto,
	}
}

func (inj *Injector) Require() map[BeanId]*Bean {
	p := &Provider{fn: inj.fn}
	bean := p.Provide()
	required := make(map[BeanId]*Bean)
	owned := make(map[BeanId]*Bean)
	for _, p := range inj.providers {
		for _, b := range p.Need() {
			required[b.String()] = b
		}
		b := p.Provide()
		owned[b.String()] = b
		if bean != nil && types.Identical(b.typ, bean.typ) {
			bean = nil
		}
	}
	if bean != nil {
		panic("at least one provider return same type as injector func result")
	}
	for beanId := range required {
		if _, ok := owned[beanId]; ok {
			delete(required, beanId)
		}
	}
	return required
}

func (inj *Injector) AddProvider(list ...*Provider) {
	for _, p := range list {
		inj.providers[p.fn.String()] = p
	}
}

func (inj *Injector) String() string {
	return inj.fn.String()
}
