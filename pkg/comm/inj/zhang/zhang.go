package zhang

import "github.com/huantedness/autowire/pkg/comm/inj/zhang/yanyan"

type Zhang struct{}

func NewZhang(yanyan yanyan.YanYan) Zhang {
	return Zhang{}
}