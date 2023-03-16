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

// hahaha
func InitShu() *Shu {
	wire.Build(NewShu, liu.NewLiu, zhao.NewZhao, yanyan.NewYanYan, guan.NewGuan, zhang.NewZhang)
	return nil
}

func InitGuan() *guan.Guan {
	wire.Build(guan.NewGuan, yanyan.NewYanYan, zhang.NewZhang)
	return nil
}
