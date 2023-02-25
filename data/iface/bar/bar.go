package bar

import "github.com/huantedness/autowire/data/iface"

type Bar /*Bar is in TypeInfo.Defs*/ struct{}

// Do implements iface.Foo
func (Bar) Do /*Do is in TypeInfo.Defs*/ (text /*text is in TypeInfo.Defs*/ string) error {
	panic("unimplemented")
}

var _ iface.Foo = (*Bar)(nil)

func Handle /*Handle is in TypeInfo.Defs*/ (foo /*foo is in TypeInfo.Defs*/ iface.Foo) {
	foo.Do("text")
}
