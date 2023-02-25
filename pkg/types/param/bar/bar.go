package bar

import "github.com/huantedness/autowire/pkg/types/param"

type Bar struct{}

// Do implements iface.Foo
func (Bar) Do(text string) error {
	panic("unimplemented")
}

var _ param.Foo = (*Bar)(nil)

func Handle(foo param.Foo, impl *param.FooImpl, bar Bar) error {
	foo.Do("text")
	impl.Do("text2")
	bar.Do("text3")
	return nil
}
