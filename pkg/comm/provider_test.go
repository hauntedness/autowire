package comm

import (
	"go/types"
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/hauntedness/autowire/conf"
)

func TestFromVar(t *testing.T) {
	pkgs, err := decorator.Load(conf.DefaultConf, "github.com/hauntedness/autowire/example/param/bar")
	if err != nil {
		t.Fatal(err)
	}
	var handle *types.Func
	for id, obj := range pkgs[0].TypesInfo.Defs {
		if id == nil || obj == nil {
			continue
		}
		if id.Name == "Handle" {
			handle = obj.(*types.Func)
		}
	}
	p := Provider{fn: handle}
	beans := p.Require()
	assertEqual(t, InterfaceKind, beans[0].Kind())
	assertEqual(t, "github.com/hauntedness/autowire/example/param", beans[0].PkgPath())
	assertEqual(t, "github.com/hauntedness/autowire/example/param.Foo", beans[0].String())
	assertEqual(t, PointerKind, beans[1].Kind())
	assertEqual(t, "github.com/hauntedness/autowire/example/param", beans[1].PkgPath())
	assertEqual(t, "*github.com/hauntedness/autowire/example/param.FooImpl", beans[1].String())
	assertEqual(t, StructKind, beans[2].Kind())
	assertEqual(t, "github.com/hauntedness/autowire/example/param/bar", beans[2].PkgPath())
	assertEqual(t, "github.com/hauntedness/autowire/example/param/bar.Bar", beans[2].String())
	// only for test purpose, error should not be a valid bean
	bean := p.Provide()
	assertEqual(t, InterfaceKind, bean.Kind())
	assertEqual(t, "", bean.PkgPath())
	assertEqual(t, "error", bean.String())
}

func assertEqual[T comparable](t *testing.T, expected T, actual T) {
	if expected != actual {
		t.Errorf("value are not equal, expected: %v, actual: %v", expected, actual)
	}
}
