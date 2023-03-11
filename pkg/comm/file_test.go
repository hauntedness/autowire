package comm

import (
	"go/types"
	"reflect"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/conf"
	"github.com/huantedness/autowire/pkg/util"
	"golang.org/x/exp/slog"
)

func Test_renamed(t *testing.T) {
	type args struct {
		bm   util.BiMap[path, alias]
		path path
		name alias
	}
	tests := []struct {
		name string
		args args
		want alias
	}{
		{
			name: "0 conflict exist, first time population",
			args: args{
				bm: func() util.BiMap[path, alias] {
					bm := util.NewBiMap[path, alias]()
					return bm
				}(),
				path: "github.com/hauntedness/autowire/pkg",
				name: "pkg",
			},
			want: "pkg",
		},
		{
			name: "normal package path, 0 conflict exist",
			args: args{
				bm: func() util.BiMap[path, alias] {
					bm := util.NewBiMap[path, alias]()
					bm.MustPut("github.com/hauntedness/autowire/pkg", "pkg")
					return bm
				}(),
				path: "github.com/hauntedness/autowire/pkg",
				name: "pkg",
			},
			want: "pkg",
		},
		{
			name: "1 conflict exists, module name doesn't contain strange charactors",
			args: args{
				bm: func() util.BiMap[path, alias] {
					bm := util.NewBiMap[path, alias]()
					bm.MustPut("github.com/hauntedness/autowire/pkg", "pkg")
					return bm
				}(),
				path: "github.com/hauntedness/others/pkg",
				name: "pkg",
			},
			want: "otherspkg",
		},
		{
			name: "1 conflict exists, but package already renamed",
			args: args{
				bm: func() util.BiMap[path, alias] {
					bm := util.NewBiMap[path, alias]()
					bm.MustPut("github.com/hauntedness/autowire/pkg", "pkg")
					bm.MustPut("github.com/hauntedness/others/pkg", "otherspkg")
					return bm
				}(),
				path: "github.com/hauntedness/others/pkg",
				name: "pkg",
			},
			want: "otherspkg",
		},
		{
			name: "2 conflicts exist, module name(or parent package path) contain strange charactors",
			args: args{
				bm: func() util.BiMap[path, alias] {
					bm := util.NewBiMap[path, alias]()
					bm.MustPut("github.com/hauntedness/autowire/pkg", "pkg")
					bm.MustPut("github.com/someotheruser1/1some$sign_here/pkg", "_some_sign_herepkg")
					return bm
				}(),
				path: "github.com/someotheruser2/1some$sign_here/pkg",
				name: "pkg",
			},
			want: "_some_sign_herepkg1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renamed(tt.args.bm, tt.args.path, tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("renamed() = %v, want %v", got, tt.want)
			}
		})
	}
}

// test wire file is hard as we should want see source file can change properly in this test
// currently we only processing the refactor and check the imports and build call args length
func TestWireFile_Refactor(t *testing.T) {
	pkg, _, injector := newShuInjector(t)
	// lack of yanyan currently
	wireFile := WireFile{
		file: pkg.Syntax[0],
		injectors: map[string]*Injector{
			injector.String(): injector,
		},
	}
	wireFile.Refactor()
}

// newShuInjector load inj.NewShu and package of inj.NewShu
func newShuInjector(t *testing.T) (injPkg *decorator.Package, guanPkg *decorator.Package, inj *Injector) {
	path := "github.com/huantedness/autowire/pkg/comm/inj"
	pkgs, err := decorator.Load(conf.DefaultConf, path)
	if err != nil {
		t.Fatal(err)
	}
	path_guan := "github.com/huantedness/autowire/pkg/comm/inj/guan"
	pkg_guan, err := decorator.Load(conf.DefaultConf, path_guan)
	if err != nil {
		t.Fatal(err)
	}
	injector := &Injector{fn: nil, origin: map[string]*Provider{}, providers: map[string]*Provider{}, auto: true}
	for id, obj := range pkgs[0].TypesInfo.Defs {
		if id == nil || obj == nil {
			continue
		}
		if id.Name == "InitShu" {
			injector.fn = obj.(*types.Func)
		}
		if id.Name == "NewShu" {
			provider := &Provider{fn: obj.(*types.Func)}
			injector.origin[provider.fn.String()] = provider
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
	// for package
	for _, f := range pkgs[0].Syntax {
		// find wire.go
		if !strings.Contains(f.Name.Name, "inj") {
			continue
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*dst.FuncDecl); ok {
				// find InitShu function
				if fn.Name.Name != "InitShu" {
					continue
				}
				for _, stmt := range fn.Body.List {
					if expr, ok := stmt.(*dst.ExprStmt); ok {
						if call, ok := expr.X.(*dst.CallExpr); ok {
							injector.buildCall = call
						}
					}
				}
			}
		}
	}
	beans := injector.Require()
	for _, bean := range beans {
		slog.Info("Need", "Id", bean)
	}
	for k := range injector.providers {
		slog.Info("owned", "funcId", k)
	}
	return pkgs[0], pkg_guan[0], injector
}
