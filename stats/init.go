package stats

import (
	"expvar"
	"sync"
)

var (
	// The reference to the Registry singleton.
	Root Registry

	container *expvar.Map
	mu        sync.Mutex
)

// Configures the top level expvar Map to be used to contain
// the replaceble "statistics" Map. Any existing registrations
// will be copied into the specified Map.
//
// Note: the 'stats' API will operate without this call being
// made. If this call isn't made, then a walk of the 'expvar' tree
// with expvar.Do() will not discover the "statistics" Map containing
// all the registered Statistics objects. By making invoking this method, the
// caller can choose where to place the "statistics" Map.
//
func Init(replacement *expvar.Map) {
	mu.Lock()
	defer mu.Unlock()

	tmp := map[string]expvar.Var{}
	container.Do(func(kv expvar.KeyValue) {
		tmp[kv.Key] = kv.Value
	})
	for k, v := range tmp {
		replacement.Set(k, v)
	}
	container = replacement
}

// Ensure that container is always defined and contains a "statistics" map.
func init() {
	Root = &registry{
		listeners: make([]*listener, 0),
	}
	container = &expvar.Map{}
	container.Init()

	stats := &expvar.Map{}
	stats.Init()
	container.Set("statistics", stats)
}
