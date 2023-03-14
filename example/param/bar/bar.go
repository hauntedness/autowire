package bar

import "github.com/huantedness/autowire/example/param"

type Bar struct{}

// Do implements iface.Foo
func (Bar) Do(text string) error {
	panic("unimplemented")
}

var _ param.Foo = (*Bar)(nil)

func Handle(foo param.Foo, impl *param.FooImpl, bar Bar) error {
	_ = foo.Do("text")
	_ = impl.Do("text2")
	err := bar.Do("text3")
	return err
}
