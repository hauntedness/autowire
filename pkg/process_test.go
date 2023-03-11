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
