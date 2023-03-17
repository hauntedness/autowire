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
	wire.Build(NewShu, zhang.NewZhang, yanyan.NewYanYan, zhao.NewZhao, liu.NewLiu, guan.NewGuan)
	return nil
}

func InitGuan() *guan.Guan {
	wire.Build(guan.NewGuan, yanyan.NewYanYan, zhang.NewZhang)
	return nil
}
