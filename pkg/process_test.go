package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDIContext_Process(t *testing.T) {
	di := NewDIContext()
	path := "github.com/huantedness/autowire/pkg/comm/inj"
	di.Process(path)

	assert.NotEmpty(t, di.objects)
	bean := di.objects[objRef{importPath: path, name: "Shu"}]
	assert.NotNil(t, bean)

	assert.NotEmpty(t, di.providers)
	provider := di.providers[objRef{importPath: path, name: "NewShu"}]
	assert.NotNil(t, provider)

	assert.NotEmpty(t, di.injectors)
	injector := di.injectors[objRef{importPath: path, name: "InitShu"}]
	assert.NotNil(t, injector)
}
