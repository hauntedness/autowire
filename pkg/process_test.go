package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDIContext_Process(t *testing.T) {
	di := NewDIContext()
	path := "github.com/huantedness/autowire/example"
	di.Process(path)

	assert.NotEmpty(t, di.objects)
	bean := di.objects[objRef{importPath: path, name: "Event"}]
	assert.NotNil(t, bean)

	assert.NotEmpty(t, di.providers)
	provider := di.providers[objRef{importPath: path, name: "NewEvent"}]
	assert.NotNil(t, provider)

	provider2 := di.providers[objRef{importPath: path + "/msg", name: "NewMessage"}]
	assert.NotNil(t, provider2)

	assert.NotEmpty(t, di.injectors)
	injector := di.injectors[objRef{importPath: path, name: "InitEvent"}]
	assert.NotNil(t, injector)
}

// test complex dependencies, to see yanyan which is in very underlayer can be enriched
func TestDIContext_Process2(t *testing.T) {
	di := NewDIContext()
	path := "github.com/huantedness/autowire/example/inj"
	di.Process(path)

	// TODO here might be a bug, please test both pointer bean and struct value bean
	assert.NotEmpty(t, di.objects)
	bean := di.objects[objRef{importPath: path, name: "Liu"}]
	assert.NotNil(t, bean)

	assert.NotEmpty(t, di.providers)
	provider := di.providers[objRef{importPath: path, name: "NewShu"}]
	assert.NotNil(t, provider)

	provider2 := di.providers[objRef{importPath: path + "/liu", name: "NewLiu"}]
	assert.NotNil(t, provider2)

	provider3 := di.providers[objRef{importPath: path + "/zhang/yanyan", name: "NewYanYan"}]
	assert.NotNil(t, provider3)

	assert.NotEmpty(t, di.injectors)
	injector := di.injectors[objRef{importPath: path, name: "InitShu"}]
	assert.NotNil(t, injector)
}
