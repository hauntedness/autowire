package pkg

import (
	"testing"

	"github.com/hauntedness/autowire/pkg/comm"
)

func TestDIContext_Process(t *testing.T) {
	di := NewDIContext(nil)
	path := "github.com/hauntedness/autowire/example"
	di.Process(path)

	assertNotEmpty[*comm.Provider](t, di.providers)
	provider := di.providers[objRef{importPath: path, name: "NewEvent"}]
	assertNotNil(t, provider)

	provider2 := di.providers[objRef{importPath: path + "/msg", name: "NewMessage"}]
	assertNotNil(t, provider2)

	assertNotEmpty[*comm.Injector](t, di.injectors)
	injector := di.injectors[objRef{importPath: path, name: "InitEvent"}]
	assertNotNil(t, injector)
}

// test complex dependencies, to see yanyan which is in very underlayer can be enriched
func TestDIContext_Process2(t *testing.T) {
	di := NewDIContext(nil)
	path := "github.com/hauntedness/autowire/example/inj"
	di.Process(path)

	assertNotEmpty[*comm.Provider](t, di.providers)
	provider := di.providers[objRef{importPath: path, name: "NewShu"}]
	assertNotNil(t, provider)

	provider2 := di.providers[objRef{importPath: path + "/liu", name: "NewLiu"}]
	assertNotNil(t, provider2)

	provider3 := di.providers[objRef{importPath: path + "/zhang/yanyan", name: "NewYanYan"}]
	assertNotNil(t, provider3)

	assertNotEmpty[*comm.Injector](t, di.injectors)
	injector := di.injectors[objRef{importPath: path, name: "InitShu"}]
	assertNotNil(t, injector)
}

func assertNotEmpty[E any, T ~[]E | ~map[objRef]E](t *testing.T, collect T) {
	if len(collect) == 0 {
		t.Fatalf("expecting not empty, got empty")
	}
}

func assertNotNil[E any, PE *E](t *testing.T, obj PE) {
	if obj == nil {
		t.Fatalf("expecting not nil, got nil")
	}
}
