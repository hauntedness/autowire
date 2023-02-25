package iface

import (
	"go/types"
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/conf"
	"golang.org/x/exp/slog"
)

func TestFoo(t *testing.T) {
	pkgs, err := decorator.Load(conf.DefaultConf, "github.com/huantedness/autowire/data/iface/bar")
	if err != nil {
		t.Fatal(err)
	}
	var foo types.Object
	var bar types.Object
	for id, obj := range pkgs[0].TypesInfo.Defs {
		if id == nil || obj == nil {
			continue
		}
		slog.Info("id", id)
		if id.Name == "foo" {
			foo = obj.Type().(*types.Named).Obj()
		}
		if id.Name == "Bar" {
			bar = obj
		}
	}
	if foo == nil {
		return
	}
	named := foo.Type().(*types.Named)
	underlying := named.Underlying()
	pointer := types.NewPointer(bar.Type())
	satisfied := types.Satisfies(pointer, underlying.(*types.Interface))
	if !satisfied {
		slog.Info("foo", "named", named, "undeerlying", underlying)
		slog.Info("bar", "bar", bar)
		t.Fail()
	}
}
