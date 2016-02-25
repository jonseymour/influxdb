package influxdb

import (
	"expvar"
	"strconv"
	"sync"

	expvar2 "github.com/influxdata/influxdb/stats/expvar"
)

var expvarMu sync.Mutex
var root *expvar2.Map // a name space for all statistics maps

func init() {
	root = &expvar2.Map{}
	root.Init()
	expvar.Publish("influx", root)
}

// NewStatistics returns an expvar-based map with the given key. Within that map
// is another map. Within there "name" is the Measurement name, "tags" are the tags,
// and values are placed at the key "values".
func NewStatistics(key, name string, tags map[string]string) *expvar.Map {
	expvarMu.Lock()
	defer expvarMu.Unlock()

	// Add expvar for this service.
	var v expvar.Var
	if v = root.Get(key); v == nil {
		v = &expvar.Map{}
		v.(*expvar.Map).Init()
		root.Set(key, v)
	}
	m := v.(*expvar.Map)

	// Set the name
	nameVar := &expvar.String{}
	nameVar.Set(name)
	m.Set("name", nameVar)

	// Set the tags
	tagsVar := &expvar.Map{}
	tagsVar.Init()
	for k, v := range tags {
		value := &expvar.String{}
		value.Set(v)
		tagsVar.Set(k, value)
	}
	m.Set("tags", tagsVar)

	// Create and set the values entry used for actual stats.
	statMap := &expvar.Map{}
	statMap.Init()
	m.Set("values", statMap)

	// Create a reference counter that we can use to release the map.
	referencesVar := &expvar.Int{}
	referencesVar.Set(2) // once for the site of usage, once for the monitor
	m.Set("references", referencesVar)
	return statMap
}

// Iterate over all the statistics maps.
func DoStatistics(fn func(expvar.KeyValue)) {
	root.Do(fn)
}

// Used to deregister a statistic when it is no longer needed.
func DeleteStatistics(key string) {
	expvarMu.Lock()
	defer expvarMu.Unlock()

	v := root.Get(key)
	if v == nil {
		return
	}

	m := v.(*expvar.Map)

	if countVar := m.Get("references"); countVar == nil {
		root.Delete(key) // if this happens, start over
		return
	} else {
		countVar.(*expvar.Int).Add(-1)
		countText := countVar.String()
		count, _ := strconv.ParseInt(countText, 10, 32)
		if count <= 0 {
			root.Delete(key)
		}
	}
}
