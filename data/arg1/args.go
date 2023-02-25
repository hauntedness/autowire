package arg1

import "github.com/huantedness/autowire/data/arg1/bar"

type Foo struct{}

func NewFoo(bar bar.Bar) Foo {
	return Foo{}
}
