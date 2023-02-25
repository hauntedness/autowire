package types

import (
	"go/types"
)

type FuncId = string

// Injector store
type Injector struct {
	fn        *types.Func
	providers map[FuncId]*Provider
}

type BeanId = string

func (inj *Injector) Need() map[BeanId]*Bean {
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
