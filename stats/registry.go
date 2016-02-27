package stats

import (
	"expvar"
	"sync"
)

// A singleton reference to the stats registry
var Root Registry = &registry{
	listeners: make([]*listener, 0),
}

// Provide an interface
type Registry interface {

	// Create a new Builder of statistics objects.
	NewBuilder(k string, n string, tags map[string]string) Builder

	// Notify listeners that a new Statistics object has been opened for the first time.
	NotifyOpen(s Statistics)

	// Register a listener for NotifyOpen events. The returned function will
	// deregister the listener when called.
	//
	// Listeners are always invoked on goroutines that currently do not hold
	// any locks over the Registry.
	OnOpen(listener func(s Openable)) func()

	// Called to iterate over the registered statistics sets
	Do(f func(s Statistics))
}

// A type used to allow callbacks to be deregistered
type listener struct {
	callback func(Openable)
	closer   func()
}

// A type used to represent a registry of all Statistics objects
type registry struct {
	mu        sync.RWMutex
	listeners []*listener
}

// Ensure the top level map is always registered.
func init() {
	m := &expvar.Map{} // this map can't be replaced because it is a top level map
	m.Init()

	r := &expvar.Map{} // this map is replaceable, since it is a value of a map
	r.Init()

	m.Set("statistics", r)
	expvar.Publish("influx", m)
}

// Cleans the registry to remove statistics that have been closed.
func (r *registry) clean() {

	// rebuild the registry map
	r.mu.Lock()
	defer r.mu.Unlock()

	cleaned := &expvar.Map{}
	cleaned.Init()
	r.do(func(stats Statistics) {
		if stats.Refs() > 0 {
			cleaned.Set(stats.Key(), stats)
		}
	})
	r.getInflux().Set("statistics", cleaned)
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
		if stats.Refs() > 0 {
			f(stats)
		}
		if stats.Refs() == 0 {
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

// get the "influx" map from expvar - this map is not replacable
func (r *registry) getInflux() *expvar.Map {
	return expvar.Get("influx").(*expvar.Map)
}

// get the "statistics" map from the "influx" map - this map is replaceable
func (r *registry) getStatistics() *expvar.Map {
	return r.getInflux().Get("statistics").(*expvar.Map)
}

// Create a new builder that retains a reference to the registry.
func (r *registry) NewBuilder(k string, n string, tags map[string]string) Builder {

	impl := &expvar.Map{}
	impl.Init()

	builder := &statistics{
		registry:   r,
		key:        k,
		name:       n,
		tags:       tags,
		impl:       impl,
		refs:       0,
		intVars:    map[string]*expvar.Int{},
		stringVars: map[string]*expvar.String{},
		floatVars:  map[string]*expvar.Float{},
	}

	return builder
}

// Used by newly opened Statistics objects to notify OnOpen
// listeners that a new Statistics object has been registered.
func (r *registry) NotifyOpen(s Statistics) {

	// clone the list of listeners
	r.mu.RLock()
	clone := make([]*listener, len(r.listeners))
	copy(clone, r.listeners)
	r.mu.RUnlock()

	// call the each of the cloned listeners without holding any lock
	for _, l := range clone {
		l.callback(s)
	}

	// update the statistics map while holding the write lock
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getStatistics().Set(s.Key(), s)

	return
}

// Register a new OnOpen listener. The listener will receive notifications for
// all open Statistics currently in the Registry and for any objects that are
// subsequently added.
func (r *registry) OnOpen(lf func(o Openable)) func() {

	existing := []Statistics{}

	// add a new listener while holding the write lock
	r.mu.Lock()
	l := &listener{
		callback: lf,
	}
	l.closer = func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		for i, e := range r.listeners {
			if e == l {
				r.listeners = append(r.listeners[:i], r.listeners[i+1:]...)
				return
			}
		}
	}

	r.do(func(s Statistics) {
		existing = append(existing, s)
	})

	r.listeners = append(r.listeners, l)
	r.mu.Unlock()

	// Call the listener on objects that were already in the map before we added a listener.
	for _, s := range existing {
		if s.Refs() > 0 {
			lf(s)
		}
	}

	// By the time we get here, the listener has received one notification for
	// each Statistics object that was in the map prior to the listener being registered
	// and one notification for each added since. The notifications won't necessarily be received
	// in order of their original delivery to other listeners.

	return l.closer
}
