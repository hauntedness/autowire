package pkg

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/hauntedness/autowire/pkg/comm"
	"golang.org/x/exp/slices"
)

// optional config when processing autowire
type ProcessConfigurer interface {
	//
	// whether to rewrite source code
	WillRewriteSource() bool

	// used to report whether a function is an injector
	// this is helpful when you only want to generate a part of functions in the package
	InjectorPredicate(fn *types.Func) bool

	// used to report whether a function can be provider
	ProviderPredicate(fn *types.Func) bool

	// used to find proper provider from multiple ones
	ProviderElect(inj *comm.Injector, bean *comm.Bean, providers map[string]*comm.Provider) *comm.Provider
}

type DefaultProcessConfigurer struct {
	willRewriteSource bool
}

// InjectorPredicate implements ProcessConfigurer
func (*DefaultProcessConfigurer) InjectorPredicate(fn *types.Func) bool {
	return true
}

// ProviderElector implements ProcessConfigurer
func (*DefaultProcessConfigurer) ProviderElect(inj *comm.Injector, bean *comm.Bean, providers map[string]*comm.Provider) *comm.Provider {
	firstSequence := make([]*comm.Provider, 0, 1)
	secondSequence := make([]*comm.Provider, 0, 1)
	for _, p := range providers {
		if p.Package() == inj.Package() {
			firstSequence = append(firstSequence, p)
		} else if len(firstSequence) == 0 {
			secondSequence = append(secondSequence, p)
		}
	}
	sortFunc := func(prev, next *comm.Provider) bool {
		if len(prev.Require()) > len(next.Require()) {
			return true
		}
		if prev.Name() < next.Name() {
			return true
		}
		// TODO maybe we should use some go directive to indicate the order
		// while current problem is go doesn't support that very well
		return false
	}
	if len(firstSequence) > 0 {
		slices.SortFunc(firstSequence, sortFunc)
		return firstSequence[0]
	}
	if len(secondSequence) > 0 {
		slices.SortFunc(secondSequence, sortFunc)
		return secondSequence[0]
	}
	panic(fmt.Errorf("no provider given")) // not reachable code or this will be treated as a bug
}

// ProviderPredicate implements ProcessConfigurer
func (*DefaultProcessConfigurer) ProviderPredicate(fn *types.Func) bool {
	return strings.HasPrefix(fn.Name(), "New")
}

// WillRewriteSource implements ProcessConfigurer
func (c *DefaultProcessConfigurer) WillRewriteSource() bool {
	return c.willRewriteSource
}

var _ ProcessConfigurer = (*DefaultProcessConfigurer)(nil)
