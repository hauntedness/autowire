package comm

import (
	"fmt"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/huantedness/autowire/pkg/util"
	"golang.org/x/exp/slog"
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

func (file *WireFile) Refactor() {
	origin, current := file.imports()
	defer file.organizeImports(origin, current)
	// for each refactor pointcut
	for key, inj := range file.injectors {
		call := inj.buildCall
		inj := file.injectors[key]
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
					call.Args = append(call.Args, dst.NewIdent(path.Name()+"."+funcName))
				}
			}
		}
	}
	// rewrite dst file with new imports
	slog.Info("package after refactor", current)
}

func (file *WireFile) organizeImports(origin util.BiMap[path, alias], current util.BiMap[path, alias]) {
	for path, alias := range current.LMap() {
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
				Value: `"` + path + `"`,
				Decs:  dst.BasicLitDecorations{},
			},
			Decs: dst.ImportSpecDecorations{},
		}
		file.file.Imports = append(file.file.Imports, spec)
	}
}

func (file *WireFile) imports() (origin util.BiMap[path, alias], current util.BiMap[path, alias]) {
	origin = util.NewBiMap[path, alias]()
	current = util.NewBiMap[path, alias]()
	for i, is := range file.file.Imports {
		var name string
		slog.Info(
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

	for i := 0; i < 255; i++ {
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
