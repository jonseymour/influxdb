package stats

import (
	"expvar"
	"sync"
)

// A singleton reference to the stats registry
var Root Registry = &registry{
	listeners: []func(Openable){},
}

// Provide an interface
type Registry interface {

	// Called to obtain a builder of new Statistics object.
	NewBuilder(k string, n string, tags map[string]string) Builder

	// Notify the registry that statistics object has been opened for the first time.
	NotifyOpen(s Statistics)

	// Register a listener for NotifyOpen events.
	OnOpen(listener func(s Openable))

	// Called to iterate over the registered statistics sets
	Do(f func(s Statistics))

	// Deprecated method.
	// Called to close a statistics set by name. Only required until
	// influxdb.CloseStatistics() is deleted.
	Close(k string)
}

// manages the registry of all statistics sets
type registry struct {
	mu        sync.RWMutex
	listeners []func(Openable)
}

// ensure the top level map is always registered
func init() {
	m := &expvar.Map{} // this map can't be replaced because it is a top level map
	m.Init()

	r := &expvar.Map{} // this map is replaceable, since it is a value of a map
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
		if set.Refs() > 0 {
			f(set)
		}
		if set.Refs() == 0 {
			closed[kv.Key] = true
		}
	})

	// rebuild the registry map
	if len(closed) > 0 {

		// if this ever becomes too expensive, we might arrange things so that the
		// cleaning happens once the number of closed entries exceeds a certain
		// number or percentage of the map size.

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

// get the "influx" map from expvar - this map is not replacable
func (r *registry) getInflux() *expvar.Map {
	return expvar.Get("influx").(*expvar.Map)
}

// get the "root" map from the "influx" map - this map is replaceable
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
		refs:       0,
		intVars:    map[string]*expvar.Int{},
		stringVars: map[string]*expvar.String{},
		floatVars:  map[string]*expvar.Float{},
	}

	return builder
}

func (r *registry) NotifyOpen(s Statistics) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, f := range r.listeners {
		f(s)
	}
	r.getRoot().Set(s.Key(), s)
	return
}

func (r *registry) OnOpen(l func(o Openable)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.listeners = append(r.listeners, l)
}

func (r *registry) Close(k string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if v := r.getRoot().Get(k); v != nil {
		v.(Statistics).Close()
	}
}
