package conf

import "golang.org/x/tools/go/packages"

var DefaultConf = &packages.Config{
	BuildFlags: []string{"-tags=wireinject"},
	Mode:       packages.NeedName | packages.NeedFiles | packages.NeedDeps | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
}
