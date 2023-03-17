package guan

import "github.com/hauntedness/autowire/example/inj/zhang"

type Guan struct{}

func NewGuan(z zhang.Zhang) *Guan {
	return &Guan{}
}
