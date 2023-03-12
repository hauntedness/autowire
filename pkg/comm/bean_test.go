package comm

import (
	"go/types"
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/huantedness/autowire/conf"
)

func TestBean_Identical(t *testing.T) {
	bean1 := prepareBean(t)
	bean2 := prepareBean(t)
	if !bean1.Identical(bean2) {
		t.Errorf(
			"beans from different loader should be identical but not, bean1: %v, bean2: %v",
			bean1,
			bean2,
		)
	}
}

func prepareBean(t *testing.T) *Bean {
	pkgs, err := decorator.Load(conf.DefaultConf, "github.com/huantedness/autowire/example/param/bar")
	if err != nil {
		t.Fatal(err)
	}
	var handle *types.Func
	for id, obj := range pkgs[0].TypesInfo.Defs {
		if id == nil || obj == nil {
			continue
		}
		if id.Name == "Handle" {
			handle = obj.(*types.Func)
		}
	}
	p := Provider{fn: handle}
	return p.Provide()
}
