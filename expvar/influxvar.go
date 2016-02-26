// Encapsulates access to the top-level influx expvar.Map
//
// The init() method of this package guarantees that the map always exist.
//
// Callers can force the initialization of the package by using the
// public Get() method.
//
// Originally created to resolve a circular dependency between the 'stats' package
// and the influxdb.NewStatistics() function.
package expvar

import (
	"expvar"
)

func init() {
	m := &expvar.Map{}
	m.Init()
	expvar.Publish("influx", m)
}

func Get() *expvar.Map {
	return expvar.Get("influx").(*expvar.Map)
}
