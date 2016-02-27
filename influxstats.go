package influxdb

import (
	"expvar"
	"github.com/influxdata/influxdb/stats"
)

// Deprecated - to be removed.
//
// Use stats.Root.NewBuilder(...).MustBuild().Open().Map() instead.
//
// NewStatistics returns an expvar-based map with the given key. Within that map
// is another map. Within there "name" is the Measurement name, "tags" are the tags,
// and values are placed at the key "values".
func NewStatistics(key, name string, tags map[string]string) *expvar.Map {
	return stats.Root.
		NewBuilder(key, name, tags).
		MustBuild().
		Open().
		Map()
}
