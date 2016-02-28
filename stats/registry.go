package stats

import (
	"expvar"
	"sync"
)

// A type used to allow callbacks to be deregistered
type listener struct {
	callback func(registration)
	closer   func()
}

// A type used to represent a registry of all Statistics objects
type registry struct {
	mu        sync.RWMutex
	listeners []*listener
	container *expvar.Map
}

// Create a new builder that retains a reference to the registry.
func (r *registry) NewBuilder(k string, tags map[string]string) Builder {
	return newBuilder(k, tags, r)
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

	cleaned := &expvar.Map{}
	cleaned.Init()
	r.do(func(stats registration) {
		if stats.refs() > 0 {
			cleaned.Set(stats.Key(), stats)
		}
	})

	r.container.Set("statistics", cleaned)
}

// Iterate over all statistics irrespective of
// whether they are closed or not and without
// any cleaning behaviour.
//
// The caller is responsible for acquiring an appropriate
// lock.
func (r *registry) do(f func(s registration)) {
	r.getStatistics().Do(func(kv expvar.KeyValue) {
		f(kv.Value.(registration))
	})
}

// get the "statistics" map from the container
func (r *registry) getStatistics() *expvar.Map {
	return r.container.Get("statistics").(*expvar.Map)
}
