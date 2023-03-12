package pkg

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/conf"
	"github.com/huantedness/autowire/pkg/comm"
	"github.com/huantedness/autowire/pkg/util"
)

// load package by path
func loadPackage(path string) *decorator.Package {
	pkgs, err := decorator.Load(conf.DefaultConf, path)
	if err != nil {
		panic(err)
	}
	return pkgs[0]
}

func (di *DIContext) loadProviderAndInjector(pkg *decorator.Package, conf *LoadConfig) {
	for id, obj := range pkg.TypesInfo.Defs {
		if id == nil || obj == nil {
			continue
		}
		ref := objRef{
			importPath: obj.Pkg().Path(),
			name:       obj.Name(),
		}
		// add objects
		di.objects[ref] = obj
		di.loadProvider(obj)

	}
	if conf.LoadMode.NeedMode(LoadInjector) {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				di.loadInjector(pkg, file, decl)
			}
		}
	}
}

// loadInjector find all injector in pkg and add them to context
func (di *DIContext) loadInjector(pkg *decorator.Package, file *dst.File, decl dst.Decl) {
	dec := pkg.Decorator
	if funcDecl, ok := dec.Ast.Nodes[decl].(*ast.FuncDecl); ok {
		callExpr, err := findInjectorBuild(pkg.TypesInfo, funcDecl)
		if err != nil {
			panic(err)
		}
		if callExpr == nil {
			// funcDecl is not a build call
			return
		}
		funcObj := pkg.TypesInfo.Defs[funcDecl.Name]
		fn := funcObj.(*types.Func)
		origin := make(map[string]*comm.Provider)
		auto := true // whether apply autowire to this injector
		for _, e := range callExpr.Args {
			obj := pkg.TypesInfo.ObjectOf(util.Unwrap(e))
			if f, ok := obj.(*types.Func); ok {
				p := comm.NewProvider(f)
				origin[f.String()] = p
				ref := objRef{importPath: p.Package(), name: p.Name()}
				if _, ok := di.providers[ref]; !ok {
					di.providers[ref] = p
				}
			} else {
				auto = false
			}
		}
		ref := objRef{
			importPath: fn.Pkg().Path(),
			name:       fn.Name(),
		}
		dstExpr := dec.Dst.Nodes[callExpr].(*dst.CallExpr)
		inj := comm.NewInjector(fn, origin, dstExpr, auto)
		di.injectors[ref] = inj
		fileName := dec.Filenames[file]
		fileRef := objRef{importPath: fn.Pkg().Path(), name: fileName}
		wireFile := di.files[fileRef]
		if wireFile == nil {
			wireFile = comm.NewFile(file)
			di.files[fileRef] = wireFile
		}
		wireFile.AddInjector(inj)
	}
}

func (di *DIContext) loadProvider(obj types.Object) {
	switch fn := obj.(type) {
	case *types.Func:
		switch obj.Type().(type) {
		case *types.Signature:
			// possibly a provider
			if strings.HasPrefix(fn.Name(), "New") {
				ref := objRef{
					importPath: obj.Pkg().Path(),
					name:       obj.Name(),
				}
				di.providers[ref] = comm.NewProvider(fn)
			}
		default:
		}
	default:
	}
	return
}
