package stats

import (
	"expvar"
)

var (
	root *registry
	// The reference to the Registry singleton.
	Root Registry
)

const (
	ConfigContainer = "container"
	ConfigKey       = "key"
)

// Configures the top level expvar Map to be used to contain
// the replaceble "statistics" Map. Any existing registrations
// will be copied into the specified Map.
//
//    topMap := &expvar.Map{}
//    topMap.Init()
//    expvar.Publish("top", topMap)
//
//    stats.Init({"container": topMap, "key": "statistics"})
//
// Note: the 'stats' API will operate without this call being
// made. If this call isn't made, then a walk of the 'expvar' tree
// with expvar.Do() will not discover the "statistics" Map containing
// all the registered Statistics objects. By making invoking this method, the
// caller can choose where to place the "statistics" Map.
//
func Init(config map[string]interface{}) {
	root.init(config)
}

// Ensure that container is always defined and contains a "statistics" map.
func init() {
	statsKey := "statistics"

	container := &expvar.Map{}
	container.Init()

	stats := &expvar.Map{}
	stats.Init()

	container.Set(statsKey, stats)
	root = &registry{
		listeners:     make([]*listener, 0),
		container:     container,
		statisticsKey: statsKey,
		config: map[string]interface{}{
			ConfigContainer: container,
			ConfigKey:       statsKey,
		},
	}

	Root = root
}
