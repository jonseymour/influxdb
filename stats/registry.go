package stats

import (
	"expvar"
	"sync"
)

// A singleton reference to the stats registry
var Root Registry = &registry{}

// Provide an interface
type Registry interface {

	// Called to obtain a builder of new Statistics object.
	NewBuilder(k string, n string, tags map[string]string) Builder

	// Called to register a successfully opened Statistics with the registry.
	Open(s Statistics)

	// Called to iterate over the registered statistics sets
	Do(f func(s Statistics))

	// Deprecated method.
	// Called to close a statistics set by name. Only required until
	// influxdb.CloseStatistics() is deleted.
	Close(k string)
}

// manages the registry of all statistics sets
type registry struct {
	mu sync.RWMutex
}

// ensure the top level map is always registered
func init() {
	m := &expvar.Map{}
	m.Init()
	r := &expvar.Map{}
	r.Init()
	m.Set("root", r)
	expvar.Publish("influx", m)
}

// iterate over the registry executing the function
// specified. after it executes, we check to see if
// the set is now closed. if it is, then we
// we rebuild the top level map.
func (r *registry) Do(f func(s Statistics)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	closed := map[string]bool{}
	r.getRoot().Do(func(kv expvar.KeyValue) {
		set := kv.Value.(Statistics)
		f(set)
		if set.Refs() == 0 {
			closed[kv.Key] = true
		}
	})

	// rebuild the registry map
	if len(closed) > 0 {
		cleaned := &expvar.Map{}
		cleaned.Init()
		r.getRoot().Do(func(kv expvar.KeyValue) {
			if !closed[kv.Key] {
				cleaned.Set(kv.Key, kv.Value)
			}
		})
		r.getInflux().Set("root", cleaned)
	}
}

// get the top level map from expvar
func (r *registry) getInflux() *expvar.Map {
	return expvar.Get("influx").(*expvar.Map)
}

// get the top level map from expvar
func (r *registry) getRoot() *expvar.Map {
	return r.getInflux().Get("root").(*expvar.Map)
}

// create a new builder
func (r *registry) NewBuilder(k string, n string, tags map[string]string) Builder {
	r.mu.Lock()
	defer r.mu.Unlock()

	impl := &expvar.Map{}
	impl.Init()

	builder := &statistics{
		key:        k,
		registry:   r,
		name:       n,
		tags:       tags,
		impl:       impl,
		refs:       1, // an implicit open on behalf of the monitor
		intVars:    map[string]*expvar.Int{},
		stringVars: map[string]*expvar.String{},
		floatVars:  map[string]*expvar.Float{},
	}

	return builder
}

func (r *registry) Open(s Statistics) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if s.Refs() >= 2 {
		// the Statistics object doesn't iterated over until ref count is at least 2
		// this prevents the monitor prematurely closing it before the first Open() call
		// has been made.
		r.getRoot().Set(s.Key(), s)
	}
	return
}

func (r *registry) Close(k string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if v := r.getRoot().Get(k); v != nil {
		v.(Statistics).Close()
	}
}
