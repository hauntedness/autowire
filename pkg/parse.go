// Copyright 2018 The Wire Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// copied from github.com/google/wire/internal/wire
package pkg

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

// A providerSetSrc captures the source for a type provided by a ProviderSet.
// Exactly one of the fields will be set.
type providerSetSrc struct {
	Provider    *Provider
	Binding     *IfaceBinding
	Value       *Value
	Import      *ProviderSet
	InjectorArg *InjectorArg
	Field       *Field
}

// description returns a string describing the source of p, including line numbers.
func (p *providerSetSrc) description(fset *token.FileSet, typ types.Type) string {
	quoted := func(s string) string {
		if s == "" {
			return ""
		}
		return fmt.Sprintf("%q ", s)
	}
	switch {
	case p.Provider != nil:
		kind := "provider"
		if p.Provider.IsStruct {
			kind = "struct provider"
		}
		return fmt.Sprintf("%s %s(%s)", kind, quoted(p.Provider.Name), fset.Position(p.Provider.Pos))
	case p.Binding != nil:
		return fmt.Sprintf("wire.Bind (%s)", fset.Position(p.Binding.Pos))
	case p.Value != nil:
		return fmt.Sprintf("wire.Value (%s)", fset.Position(p.Value.Pos))
	case p.Import != nil:
		return fmt.Sprintf("provider set %s(%s)", quoted(p.Import.VarName), fset.Position(p.Import.Pos))
	case p.InjectorArg != nil:
		args := p.InjectorArg.Args
		return fmt.Sprintf("argument %s to injector function %s (%s)", args.Tuple.At(p.InjectorArg.Index).Name(), args.Name, fset.Position(args.Pos))
	case p.Field != nil:
		return fmt.Sprintf("wire.FieldsOf (%s)", fset.Position(p.Field.Pos))
	}
	panic("providerSetSrc with no fields set")
}

// trace returns a slice of strings describing the (possibly recursive) source
// of p, including line numbers.
func (p *providerSetSrc) trace(fset *token.FileSet, typ types.Type) []string {
	var retval []string
	// Only Imports need recursion.
	if p.Import != nil {
		if parent := p.Import.srcMap.At(typ); parent != nil {
			retval = append(retval, parent.(*providerSetSrc).trace(fset, typ)...)
		}
	}
	retval = append(retval, p.description(fset, typ))
	return retval
}

// A ProviderSet describes a set of providers.  The zero value is an empty
// ProviderSet.
type ProviderSet struct {
	// Pos is the position of the call to wire.NewSet or wire.Build that
	// created the set.
	Pos token.Pos
	// PkgPath is the import path of the package that declared this set.
	PkgPath string
	// VarName is the variable name of the set, if it came from a package
	// variable.
	VarName string

	Providers []*Provider
	Bindings  []*IfaceBinding
	Values    []*Value
	Fields    []*Field
	Imports   []*ProviderSet
	// InjectorArgs is only filled in for wire.Build.
	InjectorArgs *InjectorArgs

	// providerMap maps from provided type to a *ProvidedType.
	// It includes all of the imported types.
	providerMap *typeutil.Map

	// srcMap maps from provided type to a *providerSetSrc capturing the
	// Provider, Binding, Value, or Import that provided the type.
	srcMap *typeutil.Map
}

// Outputs returns a new slice containing the set of possible types the
// provider set can produce. The order is unspecified.
func (set *ProviderSet) Outputs() []types.Type {
	return set.providerMap.Keys()
}

// For returns a ProvidedType for the given type, or the zero ProvidedType.
func (set *ProviderSet) For(t types.Type) ProvidedType {
	pt := set.providerMap.At(t)
	if pt == nil {
		return ProvidedType{}
	}
	return *pt.(*ProvidedType)
}

// An IfaceBinding declares that a type should be used to satisfy inputs
// of the given interface type.
type IfaceBinding struct {
	// Iface is the interface type, which is what can be injected.
	Iface types.Type

	// Provided is always a type that is assignable to Iface.
	Provided types.Type

	// Pos is the position where the binding was declared.
	Pos token.Pos
}

// Provider records the signature of a provider. A provider is a
// single Go object, either a function or a named struct type.
type Provider struct {
	// Pkg is the package that the Go object resides in.
	Pkg *types.Package

	// Name is the name of the Go object.
	Name string

	// Pos is the source position of the func keyword or type spec
	// defining this provider.
	Pos token.Pos

	// Args is the list of data dependencies this provider has.
	Args []ProviderInput

	// Varargs is true if the provider function is variadic.
	Varargs bool

	// IsStruct is true if this provider is a named struct type.
	// Otherwise it's a function.
	IsStruct bool

	// Out is the set of types this provider produces. It will always
	// contain at least one type.
	Out []types.Type

	// HasCleanup reports whether the provider function returns a cleanup
	// function.  (Always false for structs.)
	HasCleanup bool

	// HasErr reports whether the provider function can return an error.
	// (Always false for structs.)
	HasErr bool
}

// ProviderInput describes an incoming edge in the provider graph.
type ProviderInput struct {
	Type types.Type

	// If the provider is a struct, FieldName will be the field name to set.
	FieldName string
}

// Value describes a value expression.
type Value struct {
	// Pos is the source position of the expression defining this value.
	Pos token.Pos

	// Out is the type this value produces.
	Out types.Type

	// expr is the expression passed to wire.Value.
	expr ast.Expr

	// info is the type info for the expression.
	info *types.Info
}

// InjectorArg describes a specific argument passed to an injector function.
type InjectorArg struct {
	// Args is the full set of arguments.
	Args *InjectorArgs
	// Index is the index into Args.Tuple for this argument.
	Index int
}

// InjectorArgs describes the arguments passed to an injector function.
type InjectorArgs struct {
	// Name is the name of the injector function.
	Name string
	// Tuple represents the arguments.
	Tuple *types.Tuple
	// Pos is the source position of the injector function.
	Pos token.Pos
}

// Field describes a specific field selected from a struct.
type Field struct {
	// Parent is the struct or pointer to the struct that the field belongs to.
	Parent types.Type
	// Name is the field name.
	Name string
	// Pkg is the package that the struct resides in.
	Pkg *types.Package
	// Pos is the source position of the field declaration.
	// defining these fields.
	Pos token.Pos
	// Out is the field's provided types. The first element provides the
	// field type. If the field is coming from a pointer to a struct,
	// there will be a second element providing a pointer to the field.
	Out []types.Type
}

// load typechecks the packages that match the given patterns and
// includes source for all transitive dependencies. The patterns are
// defined by the underlying build system. For the go tool, this is
// described at https://golang.org/cmd/go/#hdr-Package_lists_and_patterns
//
// wd is the working directory and env is the set of environment
// variables to use when loading the packages specified by patterns. If
// env is nil or empty, it is interpreted as an empty set of variables.
// In case of duplicate environment variables, the last one in the list
// takes precedence.
func load(ctx context.Context, wd string, env []string, tags string, patterns []string) ([]*packages.Package, []error) {
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.LoadAllSyntax,
		Dir:        wd,
		Env:        env,
		BuildFlags: []string{"-tags=wireinject"},
		// TODO(light): Use ParseFile to skip function bodies and comments in indirect packages.
	}
	if len(tags) > 0 {
		cfg.BuildFlags[0] += " " + tags
	}
	escaped := make([]string, len(patterns))
	for i := range patterns {
		escaped[i] = "pattern=" + patterns[i]
	}
	pkgs, err := packages.Load(cfg, escaped...)
	if err != nil {
		return nil, []error{err}
	}
	var errs []error
	for _, p := range pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return pkgs, nil
}

// Info holds the result of Load.
type Info struct {
	Fset *token.FileSet

	// Sets contains all the provider sets in the initial packages.
	Sets map[ProviderSetID]*ProviderSet

	// Injectors contains all the injector functions in the initial packages.
	// The order is undefined.
	Injectors []*Injector
}

// A ProviderSetID identifies a named provider set.
type ProviderSetID struct {
	ImportPath string
	VarName    string
}

// String returns the ID as ""path/to/pkg".Foo".
func (id ProviderSetID) String() string {
	return strconv.Quote(id.ImportPath) + "." + id.VarName
}

// An Injector describes an injector function.
type Injector struct {
	ImportPath string
	FuncName   string
}

// String returns the injector name as ""path/to/pkg".Foo".
func (in *Injector) String() string {
	return strconv.Quote(in.ImportPath) + "." + in.FuncName
}

// objectCache is a lazily evaluated mapping of objects to Wire structures.
type objectCache struct {
	fset     *token.FileSet
	packages map[string]*packages.Package
	objects  map[objRef]objCacheEntry
	hasher   typeutil.Hasher
}

type objRef struct {
	importPath string
	name       string
}

type objCacheEntry struct {
	val  interface{} // *Provider, *ProviderSet, *IfaceBinding, or *Value
	errs []error
}

func newObjectCache(pkgs []*packages.Package) *objectCache {
	if len(pkgs) == 0 {
		panic("object cache must have packages to draw from")
	}
	oc := &objectCache{
		fset:     pkgs[0].Fset,
		packages: make(map[string]*packages.Package),
		objects:  make(map[objRef]objCacheEntry),
		hasher:   typeutil.MakeHasher(),
	}
	// Depth-first search of all dependencies to gather import path to
	// packages.Package mapping. go/packages guarantees that for a single
	// call to packages.Load and an import path X, there will exist only
	// one *packages.Package value with PkgPath X.
	stk := append([]*packages.Package(nil), pkgs...)
	for len(stk) > 0 {
		p := stk[len(stk)-1]
		stk = stk[:len(stk)-1]
		if oc.packages[p.PkgPath] != nil {
			continue
		}
		oc.packages[p.PkgPath] = p
		for _, imp := range p.Imports {
			stk = append(stk, imp)
		}
	}
	return oc
}

// varDecl finds the declaration that defines the given variable.
func (oc *objectCache) varDecl(obj *types.Var) *ast.ValueSpec {
	// TODO(light): Walk files to build object -> declaration mapping, if more performant.
	// Recommended by https://golang.org/s/types-tutorial
	pkg := oc.packages[obj.Pkg().Path()]
	pos := obj.Pos()
	for _, f := range pkg.Syntax {
		tokenFile := oc.fset.File(f.Pos())
		if base := tokenFile.Base(); base <= int(pos) && int(pos) < base+tokenFile.Size() {
			path, _ := astutil.PathEnclosingInterval(f, pos, pos)
			for _, node := range path {
				if spec, ok := node.(*ast.ValueSpec); ok {
					return spec
				}
			}
		}
	}
	return nil
}

// structArgType attempts to interpret an expression as a simple struct type.
// It assumes any parentheses have been stripped.
func structArgType(info *types.Info, expr ast.Expr) *types.TypeName {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	tn, ok := qualifiedIdentObject(info, lit.Type).(*types.TypeName)
	if !ok {
		return nil
	}
	if _, isStruct := tn.Type().Underlying().(*types.Struct); !isStruct {
		return nil
	}
	return tn
}

// qualifiedIdentObject finds the object for an identifier or a
// qualified identifier, or nil if the object could not be found.
func qualifiedIdentObject(info *types.Info, expr ast.Expr) types.Object {
	switch expr := expr.(type) {
	case *ast.Ident:
		return info.ObjectOf(expr)
	case *ast.SelectorExpr:
		pkgName, ok := expr.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if _, ok := info.ObjectOf(pkgName).(*types.PkgName); !ok {
			return nil
		}
		return info.ObjectOf(expr.Sel)
	default:
		return nil
	}
}

func injectorFuncSignature(sig *types.Signature) (*types.Tuple, outputSignature, error) {
	out, err := funcOutput(sig)
	if err != nil {
		return nil, outputSignature{}, err
	}
	return sig.Params(), out, nil
}

type outputSignature struct {
	out     types.Type
	cleanup bool
	err     bool
}

// funcOutput validates an injector or provider function's return signature.
func funcOutput(sig *types.Signature) (outputSignature, error) {
	results := sig.Results()
	switch results.Len() {
	case 0:
		return outputSignature{}, errors.New("no return values")
	case 1:
		return outputSignature{out: results.At(0).Type()}, nil
	case 2:
		out := results.At(0).Type()
		switch t := results.At(1).Type(); {
		case types.Identical(t, errorType):
			return outputSignature{out: out, err: true}, nil
		case types.Identical(t, cleanupType):
			return outputSignature{out: out, cleanup: true}, nil
		default:
			return outputSignature{}, fmt.Errorf("second return type is %s; must be error or func()", types.TypeString(t, nil))
		}
	case 3:
		if t := results.At(1).Type(); !types.Identical(t, cleanupType) {
			return outputSignature{}, fmt.Errorf("second return type is %s; must be func()", types.TypeString(t, nil))
		}
		if t := results.At(2).Type(); !types.Identical(t, errorType) {
			return outputSignature{}, fmt.Errorf("third return type is %s; must be error", types.TypeString(t, nil))
		}
		return outputSignature{
			out:     results.At(0).Type(),
			cleanup: true,
			err:     true,
		}, nil
	default:
		return outputSignature{}, errors.New("too many return values")
	}
}

func allFields(call *ast.CallExpr) bool {
	if len(call.Args) != 2 {
		return false
	}
	b, ok := call.Args[1].(*ast.BasicLit)
	if !ok {
		return false
	}
	return strings.EqualFold(strconv.Quote("*"), b.Value)
}

// isPrevented checks whether field i is prevented by tag "-".
// Since this is the only tag used by wire, we can do string comparison
// without using reflect.
func isPrevented(tag string) bool {
	return reflect.StructTag(tag).Get("wire") == "-"
}

// checkField reports whether f is a field of st. f should be a string with the
// field name.
func checkField(f ast.Expr, st *types.Struct) (*types.Var, error) {
	b, ok := f.(*ast.BasicLit)
	if !ok {
		return nil, fmt.Errorf("%v must be a string with the field name", f)
	}
	for i := 0; i < st.NumFields(); i++ {
		if strings.EqualFold(strconv.Quote(st.Field(i).Name()), b.Value) {
			if isPrevented(st.Tag(i)) {
				return nil, fmt.Errorf("%s is prevented from injecting by wire", b.Value)
			}
			return st.Field(i), nil
		}
	}
	return nil, fmt.Errorf("%s is not a field of %s", b.Value, st.String())
}

// findInjectorBuild returns the wire.Build call if fn is an injector template.
// It returns nil if the function is not an injector template.
func findInjectorBuild(info *types.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
	if fn.Body == nil {
		return nil, nil
	}
	numStatements := 0
	invalid := false
	var wireBuildCall *ast.CallExpr
	for _, stmt := range fn.Body.List {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			numStatements++
			if numStatements > 1 {
				invalid = true
			}
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			if qualifiedIdentObject(info, call.Fun) == types.Universe.Lookup("panic") {
				if len(call.Args) != 1 {
					continue
				}
				call, ok = call.Args[0].(*ast.CallExpr)
				if !ok {
					continue
				}
			}
			buildObj := qualifiedIdentObject(info, call.Fun)
			if buildObj == nil || buildObj.Pkg() == nil || !isWireImport(buildObj.Pkg().Path()) || buildObj.Name() != "Build" {
				continue
			}
			wireBuildCall = call
		case *ast.EmptyStmt:
			// Do nothing.
		case *ast.ReturnStmt:
			// Allow the function to end in a return.
			if numStatements == 0 {
				return nil, nil
			}
		default:
			invalid = true
		}
	}
	if wireBuildCall == nil {
		return nil, nil
	}
	if invalid {
		return nil, errors.New("a call to wire.Build indicates that this function is an injector, but injectors must consist of only the wire.Build call and an optional return")
	}
	return wireBuildCall, nil
}

func isWireImport(path string) bool {
	// TODO(light): This is depending on details of the current loader.
	const vendorPart = "vendor/"
	if i := strings.LastIndex(path, vendorPart); i != -1 && (i == 0 || path[i-1] == '/') {
		path = path[i+len(vendorPart):]
	}
	return path == "github.com/google/wire"
}

func isProviderSetType(t types.Type) bool {
	n, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := n.Obj()
	return obj.Pkg() != nil && isWireImport(obj.Pkg().Path()) && obj.Name() == "ProviderSet"
}

// ProvidedType represents a type provided from a source. The source
// can be a *Provider (a provider function), a *Value (wire.Value), or an
// *InjectorArgs (arguments to the injector function). The zero value has
// none of the above, and returns true for IsNil.
type ProvidedType struct {
	// t is the provided concrete type.
	t types.Type
	p *Provider
	v *Value
	a *InjectorArg
	f *Field
}

// IsNil reports whether pt is the zero value.
func (pt ProvidedType) IsNil() bool {
	return pt.p == nil && pt.v == nil && pt.a == nil && pt.f == nil
}

// Type returns the output type.
//
//   - For a function provider, this is the first return value type.
//   - For a struct provider, this is either the struct type or the pointer type
//     whose element type is the struct type.
//   - For a value, this is the type of the expression.
//   - For an argument, this is the type of the argument.
func (pt ProvidedType) Type() types.Type {
	return pt.t
}

// IsProvider reports whether pt points to a Provider.
func (pt ProvidedType) IsProvider() bool {
	return pt.p != nil
}

// IsValue reports whether pt points to a Value.
func (pt ProvidedType) IsValue() bool {
	return pt.v != nil
}

// IsArg reports whether pt points to an injector argument.
func (pt ProvidedType) IsArg() bool {
	return pt.a != nil
}

// IsField reports whether pt points to a Fields.
func (pt ProvidedType) IsField() bool {
	return pt.f != nil
}

// Provider returns pt as a Provider pointer. It panics if pt does not point
// to a Provider.
func (pt ProvidedType) Provider() *Provider {
	if pt.p == nil {
		panic("ProvidedType does not hold a Provider")
	}
	return pt.p
}

// Value returns pt as a Value pointer. It panics if pt does not point
// to a Value.
func (pt ProvidedType) Value() *Value {
	if pt.v == nil {
		panic("ProvidedType does not hold a Value")
	}
	return pt.v
}

// Arg returns pt as an *InjectorArg representing an injector argument. It
// panics if pt does not point to an arg.
func (pt ProvidedType) Arg() *InjectorArg {
	if pt.a == nil {
		panic("ProvidedType does not hold an Arg")
	}
	return pt.a
}

// Field returns pt as a Field pointer. It panics if pt does not point to a
// struct Field.
func (pt ProvidedType) Field() *Field {
	if pt.f == nil {
		panic("ProvidedType does not hold a Field")
	}
	return pt.f
}

// bindShouldUsePointer loads the wire package the user is importing from their
// injector. The call is a wire marker function call.
func bindShouldUsePointer(info *types.Info, call *ast.CallExpr) bool {
	// These type assertions should not fail, otherwise panic.
	fun := call.Fun.(*ast.SelectorExpr)                 // wire.Bind
	pkgName := fun.X.(*ast.Ident)                       // wire
	wireName := info.ObjectOf(pkgName).(*types.PkgName) // wire package
	return wireName.Imported().Scope().Lookup("bindToUsePointer") != nil
}
