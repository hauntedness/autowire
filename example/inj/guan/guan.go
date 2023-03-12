package guan

import "github.com/huantedness/autowire/example/inj/zhang"

type Guan struct{}

func NewGuan(z zhang.Zhang) *Guan {
	return &Guan{}
}
