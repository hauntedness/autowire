package inj

import (
	"github.com/google/wire"
	"github.com/huantedness/autowire/pkg/comm/inj/liu"
	"github.com/huantedness/autowire/pkg/comm/inj/zhao"
)

type Shu struct{}

func NewShu(liu *liu.Liu, zhao *zhao.Zhao) *Shu {
	return &Shu{}
}

func InitShu() *Shu {
	wire.Build(NewShu, liu.NewLiu)
	return nil
}
