package influxdb

import (
	"expvar"
	"github.com/influxdata/influxdb/stats"
)

// Deprecated - to be removed.
//
// Use stats.Root.NewBuilder(...).MustBuild().Open().Map() instead.
//
// NewStatistics returns an expvar-based map with the given key.
func NewStatistics(key, name string, tags map[string]string) *expvar.Map {
	clone := make(map[string]string)
	for k, e := range tags {
		clone[k] = e
	}
	clone["name"] = name

	return stats.Root.
		NewBuilder(key, clone).
		MustBuild().
		Open().
		ValuesMap()
}
