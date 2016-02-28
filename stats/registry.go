package stats

import (
	"expvar"
	"sync"

	influxexpvar "github.com/influxdata/influxdb/expvar"
)

// This type is used by the View and the Registry to manage the
// life cycle and visibility of statistics within the registry
// and the view.
type registration interface {
	Statistics
	// True if the owner has not yet closed this object.
	IsOpen() bool
	// Increment the number observers
	Observe()
	// Decrement the number of observers.
	StopObserving() int
	// The number of open references to the receiver.
	Refs() int
}

// A type used views to register observers of new registrations
// and to issuing clean hints.
type registryClient interface {
	clean()
	// Called by a Statistics implementation to register itself when it
	// is first opened.
	register(r registration)

	onOpen(lf func(o registration)) func()
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
	r.do(func(stats registration) {
		if stats.Refs() > 0 {
			cleaned.Set(stats.Key(), stats)
		}
	})
	influxexpvar.Get().Set("statistics", cleaned)
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
	return influxexpvar.Get().Get("statistics").(*expvar.Map)
}

// Create a new builder that retains a reference to the registry.
func (r *registry) NewBuilder(k string, n string, tags map[string]string) Builder {
	return newBuilder(k, n, tags, r)
}
