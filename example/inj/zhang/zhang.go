package zhang

import "github.com/hauntedness/autowire/example/inj/zhang/yanyan"

type Zhang struct{}

func NewZhang(yanyan yanyan.YanYan) Zhang {
	return Zhang{}
}
