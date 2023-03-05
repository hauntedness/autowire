package guan

import "github.com/huantedness/autowire/pkg/comm/inj/zhang"

type Guan struct{}

func NewGuan(z zhang.Zhang) *Guan {
	return &Guan{}
}
