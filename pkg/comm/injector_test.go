package comm

import (
	"go/types"
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/hauntedness/autowire/conf"
	"golang.org/x/exp/slog"
)

func TestInjector_Need(t *testing.T) {
	path := "github.com/hauntedness/autowire/example/inj"
	pkgs, err := decorator.Load(conf.DefaultConf, path)
	if err != nil {
		t.Fatal(err)
	}
	path_guan := "github.com/hauntedness/autowire/example/inj/guan"
	pkg_guan, err := decorator.Load(conf.DefaultConf, path_guan)
	if err != nil {
		t.Fatal(err)
	}
	injector := Injector{fn: nil, providers: map[string]*Provider{}}
	for id, obj := range pkgs[0].TypesInfo.Defs {
		if id == nil || obj == nil {
			continue
		}
		if id.Name == "InitShu" {
			injector.fn = obj.(*types.Func)
		}
		if id.Name == "NewShu" {
			provider := &Provider{fn: obj.(*types.Func)}
			injector.AddProvider(provider)
		}
	}
	for id, obj := range pkgs[0].TypesInfo.Uses {
		if id == nil || obj == nil {
			continue
		}
		if id.Name == "NewLiu" {
			provider := &Provider{fn: obj.(*types.Func)}
			injector.AddProvider(provider)
		}
	}
	for id, obj := range pkg_guan[0].TypesInfo.Defs {
		if id == nil || obj == nil {
			continue
		}
		if id.Name == "NewGuan" {
			provider := &Provider{fn: obj.(*types.Func)}
			injector.AddProvider(provider)
		}
	}
	beans := injector.Require()
	for _, bean := range beans {
		slog.Info("Need", "Id", bean)
	}
	for k := range injector.providers {
		slog.Info("owned", "funcI", k)
	}
	// lack of yanyan currently
}
