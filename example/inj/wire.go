//go:build wireinject

package inj

import (
	"github.com/google/wire"
	"github.com/hauntedness/autowire/example/inj/guan"
	"github.com/hauntedness/autowire/example/inj/liu"
	"github.com/hauntedness/autowire/example/inj/zhang"
	"github.com/hauntedness/autowire/example/inj/zhang/yanyan"
	"github.com/hauntedness/autowire/example/inj/zhao"
)

type Shu struct{}

func NewShu(liu *liu.Liu, zhao *zhao.Zhao) *Shu {
	return &Shu{}
}

func InitShu() *Shu {
	wire.Build(NewShu, yanyan.NewYanYan, liu.NewLiu, zhao.NewZhao, guan.NewGuan, zhang.NewZhang)
	return nil
}
