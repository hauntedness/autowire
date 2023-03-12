package comm

import (
	"fmt"
	"go/types"

	"github.com/dave/dst"
)

type FuncId = string

// Injector store
type Injector struct {
	fn        *types.Func
	origin    map[FuncId]*Provider // original providers from the injector source code
	providers map[FuncId]*Provider // all providers after analyzed
	auto      bool                 // whether autowire can fill up this provider
	buildCall *dst.CallExpr        // the syntax tree node on which all providers lying
}

type BeanId = string

// NewInjector
func NewInjector(fn *types.Func, origin map[FuncId]*Provider, buildCall *dst.CallExpr, auto bool) *Injector {
	copy := make(map[FuncId]*Provider)
	for k, p := range origin {
		copy[k] = p
	}
	return &Injector{
		fn:        fn,
		origin:    origin,
		providers: copy,
		auto:      auto,
		buildCall: buildCall,
	}
}

// Require report beans needed to found to complete the injection
//
//	if provider P in Injector I provide Bean B, then I does not require B
//	else if No provider provide B, the injector I require B.
func (inj *Injector) Require() map[BeanId]*Bean {
	p := &Provider{fn: inj.fn}
	bean := p.Provide()
	required := make(map[BeanId]*Bean)
	owned := make(map[BeanId]*Bean)
	for _, p := range inj.providers {
		for _, b := range p.Require() {
			required[b.String()] = b
		}
		b := p.Provide()
		owned[b.String()] = b
		if bean != nil && b.Identical(bean) {
			bean = nil
		}
	}
	if bean != nil {
		err := fmt.Errorf("error: the injector func need %v, the corresponding provider should be in wire.Build func", bean.String())
		panic(err)
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
