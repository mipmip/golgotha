package provider

import (
	"fmt"
	"sync"

	"github.com/mipmip/golgotha/internal/config"
)

// Constructor builds a Provider from its configuration.
type Constructor func(p *config.Provider) (Provider, error)

// Registry maps provider types to their constructors. The zero value is not
// usable; obtain one via NewRegistry.
type Registry struct {
	mu           sync.RWMutex
	constructors map[config.ProviderType]Constructor
}

// NewRegistry returns an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{constructors: make(map[config.ProviderType]Constructor)}
}

// Register associates a constructor with a provider type. It panics if a
// constructor is already registered for the type, which surfaces duplicate
// registrations at program start.
func (r *Registry) Register(t config.ProviderType, c Constructor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.constructors[t]; exists {
		panic(fmt.Sprintf("provider: constructor already registered for type %q", t))
	}
	r.constructors[t] = c
}

// Build constructs the Provider for the given configuration, selecting the
// constructor by provider type. It returns an error for unknown types.
func (r *Registry) Build(p *config.Provider) (Provider, error) {
	r.mu.RLock()
	c, ok := r.constructors[p.Type]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no provider client registered for type %q", p.Type)
	}
	return c(p)
}
