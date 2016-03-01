package stats

import (
	"sync"
)

// A type used to allow callbacks to be deregistered
type listener struct {
	callback func(registration)
	closer   func()
}

// A type used to represent a registry of all Statistics objects
type registry struct {
	mu            sync.RWMutex
	listeners     []*listener
	registrations map[string]registration
}

// Create a new builder that retains a reference to the registry.
func (r *registry) NewBuilder(k string, n string, tags map[string]string) Builder {
	return newBuilder(k, n, tags, r)
}

func newRegistry() *registry {
	return &registry{
		listeners:     make([]*listener, 0),
		registrations: map[string]registration{},
	}
}

// Open a new view over the contents of the registry.
func (r *registry) Open() View {
	return newView(r)
}

// Cleans the registry to remove statistics that have been closed.
func (r *registry) clean() {

	// rebuild the registry map
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, g := range r.registrations {
		if g.refs() == 0 {
			delete(r.registrations, k)
		}
	}
}
