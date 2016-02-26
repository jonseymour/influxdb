package stats

import (
	"expvar"
	"sync"

	influxexpvar "github.com/influxdata/influxdb/expvar"
)

// A singleton reference to the stats registry
var Root Registry = &registry{
	listeners: make([]*listener, 0),
}

// A filter which can be used to get all statistics.
var AllStatistics = func(s Statistics) bool {
	return true
}

// Provide an interface
type Registry interface {

	// Create a new Builder of statistics objects.
	NewBuilder(k string, n string, tags map[string]string) Builder

	// Called by a Statistics implementation to register itself when it
	// is first opened.
	Register(r Registration)

	// Register a listener for NotifyOpen events. The returned function will
	// deregister the listener when called.
	//
	// Listeners are always invoked on goroutines that currently do not hold
	// any locks over the Registry.
	OnOpen(listener func(s Observable)) func()

	// Called to iterate over the registered statistics sets.
	Do(f func(s Statistics))

	// Creates a filtered copy of statistics that match the specified filter.
	Filter(filter func(s Statistics) bool) []Statistics
}

// A type used to allow callbacks to be deregistered
type listener struct {
	callback func(Observable)
	closer   func()
}

// A type used to represent a registry of all Statistics objects
type registry struct {
	mu        sync.RWMutex
	listeners []*listener
}

// Ensure the top level map is always registered.
func init() {

	r := &expvar.Map{} // this map is replaceable, since it is a value of a map
	r.Init()

	influxexpvar.Get().Set("statistics", r)
}

// Cleans the registry to remove statistics that have been closed.
func (r *registry) clean() {

	// rebuild the registry map
	r.mu.Lock()
	defer r.mu.Unlock()

	cleaned := &expvar.Map{}
	cleaned.Init()
	r.do(func(stats Statistics) {
		if stats.(Registration).Refs() > 0 {
			cleaned.Set(stats.Key(), stats)
		}
	})
	influxexpvar.Get().Set("statistics", cleaned)
}

//
// Iterates over the registry, holding a read lock.
//
// The iteration skips over closed statistics.
//
// If any closed statistics are detected during
// the operation, then the "statistics" map is
// cleansed by creating a new map and copying
// only those statistics that are still open.
//
func (r *registry) Do(f func(s Statistics)) {

	count := 0
	r.mu.RLock()
	r.do(func(stats Statistics) {
		if stats.(Registration).Refs() > 0 {
			f(stats)
		}
		if stats.(Registration).Refs() == 0 {
			count++
		}
	})
	r.mu.RUnlock()

	if count > 0 {
		r.clean()
	}
}

// Iterate over all statistics irrespective of
// whether they are closed or not and without
// any cleaning behaviour.
//
// The caller is responsible for acquiring an appropriate
// lock.
func (r *registry) do(f func(s Statistics)) {
	r.getStatistics().Do(func(kv expvar.KeyValue) {
		f(kv.Value.(Statistics))
	})
}

// Add support for returning a filtered collection of Statistics
func (r *registry) Filter(f func(s Statistics) bool) []Statistics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := []Statistics{}
	r.Do(func(s Statistics) {
		if f(s) {
			results = append(results, s)
		}
	})
	return results
}

// get the "statistics" map from the "influx" map - this map is replaceable
func (r *registry) getStatistics() *expvar.Map {
	return influxexpvar.Get().Get("statistics").(*expvar.Map)
}

// Create a new builder that retains a reference to the registry.
func (r *registry) NewBuilder(k string, n string, tags map[string]string) Builder {
	return newBuilder(k, n, tags, r)
}
