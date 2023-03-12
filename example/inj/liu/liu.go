package liu

import "github.com/huantedness/autowire/example/inj/guan"

type Liu struct{}

func NewLiu(guan *guan.Guan) *Liu {
	return &Liu{}
}
