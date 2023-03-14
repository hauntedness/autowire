package comm

import (
	"fmt"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/huantedness/autowire/logs"
	"github.com/huantedness/autowire/pkg/util"
	"golang.org/x/exp/slices"
)

type (
	path     = string
	alias    = string
	injectId = string
)

type WireFile struct {
	file      *dst.File
	injectors map[injectId]*Injector
}

func NewFile(dstFile *dst.File) *WireFile {
	return &WireFile{
		file:      dstFile,
		injectors: map[string]*Injector{},
	}
}

func (file *WireFile) AddInjector(list ...*Injector) {
	for _, inj := range list {
		file.injectors[inj.fn.String()] = inj
	}
}

func (file *WireFile) Imports() []*dst.ImportSpec {
	return file.file.Imports
}

func (file *WireFile) Refactor() {
	origin, current := file.collectImports()
	defer func() {
		// rewrite dst file with new imports
		file.organizeImports(origin, current)
		logs.Debug("package refactored", "origin imports", origin, "current imports", current)
	}()
	// for each refactor pointcut
	for _, inj := range file.injectors {
		call := inj.buildCall
		// if injector need to be refactored
		if inj.auto && len(inj.origin) < len(inj.providers) {
			for k, p := range inj.providers {
				if inj.origin[k] == nil {
					// resolve import path
					path := p.fn.Pkg()
					maybeImport(current, path)
					// resolve build call
					funcName := p.fn.Name()
					// add to call expr
					ident := dst.NewIdent(funcName)
					ident.Path = path.Path()
					call.Args = append(call.Args, ident)
				}
			}
		}
	}
}

func (file *WireFile) organizeImports(origin util.BiMap[path, alias], current util.BiMap[path, alias]) {
	// find import spec position
	var importDecl *dst.GenDecl
Declarations:
	for _, d := range file.file.Decls {
		if decl, ok := d.(*dst.GenDecl); ok {
			if _, ok := decl.Specs[0].(*dst.ImportSpec); ok {
				importDecl = decl
				break Declarations
			}
		}
	}
	type pair [2]string
	var pairs []pair
	for k, v := range current.LMap() {
		pairs = append(pairs, pair{k, v})
	}
	slices.SortFunc(pairs, func(p1, p2 pair) bool {
		return p1[0] < p2[0]
	})
	for _, v := range pairs {
		path := v[0]
		alias := v[1]
		if _, ok := origin.GetByL(path); ok {
			continue
		}
		s := strings.Split(path, "/")
		var ident *dst.Ident
		if s[len(s)-1] != alias {
			ident = dst.NewIdent(alias)
		}
		spec := &dst.ImportSpec{
			Name: ident,
			Path: &dst.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(path),
				Decs:  dst.BasicLitDecorations{},
			},
			Decs: dst.ImportSpecDecorations{},
		}
		importDecl.Specs = append(importDecl.Specs, spec)
		file.file.Imports = append(file.file.Imports, spec)
	}
}

func (file *WireFile) collectImports() (origin util.BiMap[path, alias], current util.BiMap[path, alias]) {
	origin = util.NewBiMap[path, alias]()
	current = util.NewBiMap[path, alias]()
	for i, is := range file.file.Imports {
		var name string
		logs.Info(
			"import path",
			"index", i,
			"name", is.Name,
			"path", is.Path,
			"kind", is.Path.Kind,
			"value", is.Path.Value,
		)
		path := strings.Trim(is.Path.Value, `"`)
		if is.Name == nil {
			array := strings.Split(path, "/")
			name = array[len(array)-1]
		} else {
			name = is.Name.Name
		}
		origin.Put(path, name)
		current.Put(path, name)
	}
	return origin, current
}

func maybeImport(bm util.BiMap[path, alias], pkg *types.Package) {
	// rename if conflict package alias
	alias := renamed(bm, pkg.Path(), pkg.Name())
	bm.Put(pkg.Path(), alias)
}

// renamed
func renamed(bm util.BiMap[path, alias], path path, name alias) alias {
	//

	// if path is already mapped to alias
	pkgAlias, ok := bm.GetByL(path)
	if ok {
		return pkgAlias
	}

	// if no one use the this name
	if _, ok = bm.GetByR(name); !ok {
		return name
	}

	words := strings.Split(path, "/")
	var secondLast []byte
	if length := len(words); length > 1 {
		secondLast = []byte(strings.ToLower(words[length-2]))
		for i := range secondLast {
			if secondLast[i] >= 'a' && secondLast[i] <= 'z' || i > 0 && secondLast[i] >= '0' && secondLast[i] <= '9' {
				continue
			} else {
				secondLast[i] = '_'
			}
		}
	}

	for i := range [255]struct{}{} {
		newAlias := append([]byte(string(secondLast)), name...)
		if i > 0 {
			newAlias = append(newAlias, []byte(strconv.Itoa(i))...)
		}
		if _, ok := bm.GetByR(alias(newAlias)); !ok {
			return alias(newAlias)
		}
	}
	panic(fmt.Errorf("fail to rename path after trying many times: %s", path))
}
