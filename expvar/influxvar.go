// Encapsulates access to the top-level influx expvar.Map
//
// The init() method of this package guarantees that the map always exist.
//
// Callers can force the initialization of the package by using the
// public Get() method.
//
// This package was originally created to resolve an illegal circular dependency
// between the 'influxdb/stats/registry.go' module and
// the 'influxdb/influxvar.go' module.
//
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
