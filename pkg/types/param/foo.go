package param

type Foo interface {
	Do(text string) error
}

type FooImpl struct {
	Name string
}

func (*FooImpl) Do(text string) error {
	return nil
}
