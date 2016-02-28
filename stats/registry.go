package stats

import (
	"expvar"
	"sync"
)

// A type used views to register observers of new registrations
// and to issuing clean hints.
type registryClient interface {
	// Called by a view object to register a listener for new registrations.
	onOpen(lf func(o registration)) func()
	// Called by a statistics object to register itself upon Open.
	register(r registration)
	// Called by the client when it detects that cleaning is required.
	clean()
}

// A type used to allow callbacks to be deregistered
type listener struct {
	callback func(registration)
	closer   func()
}

// A type used to represent a registry of all Statistics objects
type registry struct {
	mu        sync.RWMutex
	listeners []*listener
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

	mu.Lock()
	defer mu.Unlock()
	container.Set("statistics", cleaned)
}

func (r *registry) Open() View {
	return newView(r)
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

// get the "statistics" map from the "influx" map - this map is replaceable
func (r *registry) getStatistics() *expvar.Map {
	mu.Lock()
	defer mu.Unlock()
	return container.Get("statistics").(*expvar.Map)
}

// Create a new builder that retains a reference to the registry.
func (r *registry) NewBuilder(k string, tags map[string]string) Builder {
	return newBuilder(k, tags, r)
}
