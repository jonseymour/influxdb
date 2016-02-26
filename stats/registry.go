package stats

import (
	"expvar"
	"sync"
)

// A singleton reference to the stats registry
var Root Registry = &registry{
	listeners: make([]*listenerTicket, 0),
}

// Provide an interface
type Registry interface {

	// Called to obtain a builder of new Statistics object.
	NewBuilder(k string, n string, tags map[string]string) Builder

	// Notify the registry that statistics object has been opened for the first time.
	NotifyOpen(s Statistics)

	// Register a listener for NotifyOpen events. The returned function will
	// deregister the listener when called.
	OnOpen(listener func(s Openable)) func()

	// Called to iterate over the registered statistics sets
	Do(f func(s Statistics))
}

// manages the registry of all statistics sets
type registry struct {
	mu        sync.RWMutex
	listeners []*listenerTicket
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

	closed := map[string]struct{}{}
	r.getRoot().Do(func(kv expvar.KeyValue) {
		set := kv.Value.(Statistics)
		if set.Refs() > 0 {
			f(set)
		}
		if set.Refs() == 0 {
			closed[kv.Key] = struct{}{}
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
			if _, closed := closed[kv.Key]; !closed {
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

	for _, t := range r.listeners {
		t.callback(s)
	}
	r.getRoot().Set(s.Key(), s)
	return
}

func (r *registry) OnOpen(l func(o Openable)) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	ticket := &listenerTicket{
		callback: l,
	}
	ticket.closer = func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		for i, e := range r.listeners {
			if e == ticket {
				r.listeners = append(r.listeners[:i], r.listeners[i+1:]...)
				return
			}
		}

	}
	r.listeners = append(r.listeners, ticket)
	return ticket.closer
}

type listenerTicket struct {
	callback func(Openable)
	closer   func()
}
