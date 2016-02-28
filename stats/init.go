package stats

import (
	"expvar"
)

var (
	root *registry
	// The reference to the Registry singleton.
	Root Registry
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
func Init(config map[string]interface{}) {
	root.mu.Lock()
	defer root.mu.Unlock()

	if replacement, ok := config["container"]; ok {
		if replacement, ok := replacement.(*expvar.Map); ok {
			tmp := map[string]expvar.Var{}
			root.container.Do(func(kv expvar.KeyValue) {
				tmp[kv.Key] = kv.Value
			})
			for k, v := range tmp {
				replacement.Set(k, v)
			}
			root.container = replacement
		}
		panic("container has wrong type")
	}
}

// Ensure that container is always defined and contains a "statistics" map.
func init() {
	container := &expvar.Map{}
	container.Init()

	stats := &expvar.Map{}
	stats.Init()

	container.Set("statistics", stats)

	root = &registry{
		listeners: make([]*listener, 0),
		container: container,
	}
	Root = root
}
