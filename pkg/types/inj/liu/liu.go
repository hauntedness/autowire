package liu

import "github.com/huantedness/autowire/pkg/types/inj/guan"

type Liu struct{}

func NewLiu(guan *guan.Guan) *Liu {
	return &Liu{}
}
