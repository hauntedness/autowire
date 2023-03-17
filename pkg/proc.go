package pkg

import (
	"go/ast"
	"go/build/constraint"
	"go/types"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/hauntedness/autowire/conf"
	"github.com/hauntedness/autowire/pkg/comm"
	"github.com/hauntedness/autowire/pkg/util"
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
	// TODO check wireinjector build tag
	dec := pkg.Decorator
	astFile := dec.Ast.Nodes[file].(*ast.File)
	if ok := checkBuildConstraint(astFile); !ok {
		return
	}

	decl.Decorations()
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
			wireFile = comm.NewFile(file, fn.Pkg().Path())
			di.files[fileRef] = wireFile
		}
		wireFile.AddInjector(inj)
	}
}

func (di *DIContext) loadProvider(obj types.Object) {
	switch fn := obj.(type) {
	case *types.Func:
		switch sig := obj.Type().(type) {
		case *types.Signature:
			// just possibly a provider, so use wire validate the func
			_, err := funcOutput(sig)
			if err != nil {
				return
			}
			// and of course, we have our own validate func
			if !di.conf.ProviderPredicate(fn) {
				return
			}
			ref := objRef{
				importPath: obj.Pkg().Path(),
				name:       obj.Name(),
			}
			di.providers[ref] = comm.NewProvider(fn)
		default:
		}
	default:
	}
}

func checkBuildConstraint(file *ast.File) bool {
	var comments []string
	if len(file.Comments) > 0 {
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				comments = append(comments, c.Text)
			}
		}
	}
	if file.Doc != nil {
		for _, comment := range file.Doc.List {
			comments = append(comments, comment.Text)
		}
	}
	ok := false
	// only process go:build
	for i := range comments {
		fileTag, err := constraint.Parse(comments[i])
		if err != nil {
			continue
		}
		fileTag.Eval(func(tag string) bool {
			if tag == "wireinject" {
				ok = true
			}
			return true
		})
	}

	return ok
}
