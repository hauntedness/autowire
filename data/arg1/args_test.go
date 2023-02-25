package arg1_test

import (
	"go/types"
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/conf"
	"golang.org/x/exp/slog"
)

func TestGetPackagePathOfArgumentInOtherPackage(t *testing.T) {
	pkgs, err := decorator.Load(conf.DefaultConf, "github.com/huantedness/autowire/data/arg1")
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]
	for k, obj := range pkg.TypesInfo.Defs {
		// current pkg
		if obj == nil {
			slog.Info("key", k)
			continue
		}
		// structs , functions
		switch typ := obj.Type().(type) {
		case *types.Signature:
			s := typ.String()
			slog.Info("func", s)
		case *types.Named:
			slog.Info("Named", typ)
			fn := typ.Obj().Name()
			path := typ.Obj().Pkg().Path()
			slog.Info("obj", "name", fn, "package", path)
		case *types.Struct:
			slog.Info("Struct", typ)
		}
	}
}
