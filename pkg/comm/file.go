package comm

import (
	"fmt"
	"go/types"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/huantedness/autowire/pkg/util"
	"golang.org/x/exp/slog"
)

type WireFile struct {
	File       *dst.File
	BuildCalls map[string]*dst.CallExpr
	Injectors  map[string]*Injector
}
type (
	alias string
	path  string
)

func (file *WireFile) Refactor() {
	bm := util.NewBiDirectionMap[path, alias]()
	for _, is := range file.File.Imports {
		dst.Print(is)
	}
	// for each refactor pointcut
	for key, call := range file.BuildCalls {
		inj := file.Injectors[key]
		// if injector need to be refactored
		if inj.auto && len(inj.origin) < len(inj.providers) {
			for k, p := range inj.providers {
				if inj.origin[k] == nil {
					// resolve import path
					resolvePackage(bm, p.fn.Pkg())
					// resolve build call
					funcName := p.fn.FullName()
					call.Args = append(call.Args, dst.NewIdent(funcName))
				}
			}
		}
	}
	slog.Info("package after refactor", bm)
}

func resolvePackage(bm util.BiDirectionMap[path, alias], pkg *types.Package) {
	path := path(pkg.Path())
	alias := alias(pkg.Name())
	// rename if conflict package alias
	alias = renamed(bm, path, alias)
	bm.Put(path, alias)
}

// renamed
func renamed(bm util.BiDirectionMap[path, alias], path path, name alias) alias {
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

	words := strings.Split(string(path), "/")
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
	panic(fmt.Errorf("fail to rename path: %s", path))
}
