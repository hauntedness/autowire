package comm

import (
	"go/types"
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/conf"
	"github.com/stretchr/testify/assert"
)

func TestFromVar(t *testing.T) {
	pkgs, err := decorator.Load(conf.DefaultConf, "github.com/huantedness/autowire/example/param/bar")
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
	assert.Equal(t, InterfaceKind, beans[0].Kind())
	assert.Equal(t, "github.com/huantedness/autowire/example/param", beans[0].PkgPath())
	assert.Equal(t, "github.com/huantedness/autowire/example/param.Foo", beans[0].String())
	assert.Equal(t, PointerKind, beans[1].Kind())
	assert.Equal(t, "github.com/huantedness/autowire/example/param", beans[1].PkgPath())
	assert.Equal(t, "*github.com/huantedness/autowire/example/param.FooImpl", beans[1].String())
	assert.Equal(t, StructKind, beans[2].Kind())
	assert.Equal(t, "github.com/huantedness/autowire/example/param/bar", beans[2].PkgPath())
	assert.Equal(t, "github.com/huantedness/autowire/example/param/bar.Bar", beans[2].String())
	// only for test purpose, error should not be a valid bean
	bean := p.Provide()
	assert.Equal(t, InterfaceKind, bean.Kind())
	assert.Equal(t, "", bean.PkgPath())
	assert.Equal(t, "error", bean.String())
}
