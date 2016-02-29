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
	mu            sync.RWMutex
	listeners     []*listener
	container     *expvar.Map
	statisticsKey string
	config        map[string]interface{}
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

	r.container.Set(r.statisticsKey, cleaned)
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
	return r.container.Get(r.statisticsKey).(*expvar.Map)
}

// Initialise the registry, but don't lose any existing
// statistics.
//
// The purpose of this method is to allow the caller of Init
// to reconfigure where in the expvar tree the
// statistics objects sit.
func (r *registry) init(config map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existingStats := r.container.Get(r.statisticsKey).(*expvar.Map)

	// clone the existing configuration
	newConfig := map[string]interface{}{}
	for k, v := range config {
		newConfig[k] = v
	}

	if replacement, ok := config[ConfigContainer]; ok {
		if replacement, ok := replacement.(*expvar.Map); ok {
			var newKey = r.statisticsKey
			if tmp, ok := config[ConfigKey]; ok {
				if tmp, ok := tmp.(string); ok {
					newKey = tmp
				} else {
					panic("'key' has wrong type")
				}
			}

			replacement.Set(newKey, existingStats)

			r.container = replacement
			r.statisticsKey = newKey
		} else {
			panic("'container' has wrong type")
		}
	} else {
		// if we don't replace the container, we don't replace the key
		newConfig[ConfigKey] = r.statisticsKey
	}

	r.config = newConfig
}
