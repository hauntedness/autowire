# Autowire

Autowire is a code generation tool for easy use of [google wire][]

Wire is a code generation tool that automates connecting components using
[dependency injection][]. wire is useful but sometimes you still need to do some handwriting. 
this tool aims to simplify your work by automatically complete your source code


[dependency injection]: https://en.wikipedia.org/wiki/Dependency_injection
[google wire]: https://godoc.org/github.com/google/wire

## Installing

Install autowire by running:

```shell
go install github.com/hauntedness/autowire/cmd/autowire@latest
```

## Usage

Move to target package

```shell
cd ~/projects/autowire/example/inj
```

Make sure proper build tag and the entry provider exist

```go
//go:build wireinject

package inj

import (
	"github.com/google/wire"
	"github.com/huantedness/autowire/example/inj/liu"
	"github.com/huantedness/autowire/example/inj/zhao"
)

type Shu struct{}

func NewShu(liu *liu.Liu, zhao *zhao.Zhao) *Shu {
	return &Shu{}
}

func InitShu() *Shu {
	wire.Build(NewShu)
	return nil
}
```

Run autowire

```shell
autowire 
```
Or run go tool

```shell
go run github.com/hauntedness/autowire/cmd/autowire@latest
```

Now you should see the code is refacted.

```go
//go:build wireinject

package inj

import (
	"github.com/google/wire"
	"github.com/huantedness/autowire/example/inj/guan"
	"github.com/huantedness/autowire/example/inj/liu"
	"github.com/huantedness/autowire/example/inj/zhang"
	"github.com/huantedness/autowire/example/inj/zhang/yanyan"
	"github.com/huantedness/autowire/example/inj/zhao"
)

type Shu struct{}

func NewShu(liu *liu.Liu, zhao *zhao.Zhao) *Shu {
	return &Shu{}
}

func InitShu() *Shu {
	wire.Build(NewShu, yanyan.NewYanYan, liu.NewLiu, zhao.NewZhao, guan.NewGuan, zhang.NewZhang)
	return nil
}
```

## Note

Current limitation

- The code completion only works for the function provider, A workaround is manually create a function 
- By default, autowire only treat functions like NewXXX() bean as a valid provider
- Autowire also have a default algrithim to pick provider from multiple matches.
- If the default behavior is not what you need, you can replace it with your own implementation. see github.com/huantedness/autowire/pkg.ProcessConfigurer.