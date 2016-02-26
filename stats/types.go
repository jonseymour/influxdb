package stats

import (
	"expvar"
)

// This is the type produced by the Builder.Build() and Builder.MustBuild() methods
type Built interface {
	Open() Owner
}

// This type is used by observers to read the state of the Statistics object.
type Statistics interface {
	expvar.Var

	// The statistic key
	Key() string
	// The statistic name
	Name() string
	// The statistic tags
	Tags() map[string]string
	// The underlying values map
	Map() *expvar.Map
	// A raw values map
	Values() map[string]interface{}

	// True if the owner has not yet closed this object.
	IsOpen() bool

	// Increment the reference count for one observer.
	Observe() Statistics

	// Undoes a previous Observe.
	StopObserving()
}

// This type guarantees that observers obtain a reference to a Statistics
// object before using it.
type Observable interface {
	Key() string
	Observe() Statistics
}

// This type is used by the Registry to decide when it can drop its
// reference to a Statistics object.
type Registration interface {
	Statistics
	// The number of open references to the receiver.
	Refs() int
}

// The type is used by the owner of the Statistics object to update it and
// to close it when done.
type Owner interface {
	Statistics
	// Set a level statistics to a particular integer value
	SetInt(n string, i int64) Owner
	// Set a level statistic to a particular float value
	SetFloat(n string, f float64) Owner
	// Set a string statistic
	SetString(n string, s string) Owner

	// Add an int value to an int statistic
	AddInt(n string, i int64) Owner
	// Add a float value to a float statistic
	AddFloat(n string, f float64) Owner
	// Drop one reference to the object obtained with Open(). T
	// The described object closes the Statistics object first, then the monitor.
	Close()
}
