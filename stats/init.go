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
func Init(replacement *expvar.Map) {
	mu.Lock()
	defer mu.Unlock()

	container.Do(func(kv expvar.KeyValue) {
		replacement.Set(kv.Key, kv.Value)
	})
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
