package stats

import (
	"expvar"
)

// The types defined in this module reflect different states or roles of a Statistics objects.
//
// Consumers of Statistics objects get exactly the interface they need to perform
// the task their role requires and no more.
//
// For example, consider the life cycle of a statistics object.
//
// It starts out as a Builder, until it is built when it becomes Built. To obtain
// access to the runtime interface, the owner must execute Built.Open() which yields
// an Recorder. This provides access to the Set and Add methods and a Close() method
// to be called when done.
//
// The Registry needs to keep track of which registrations are currently referenced to
// do this, it uses the Registration interface which extends Statistics and provides
// methods to register and deregister observers and keep track of the number of references.
//

// This interface is used to declare the statistics fields during initialisation
// of a Statistics interface. A builder may be used at most once.
type Builder interface {
	DeclareInt(n string, iv int64) Builder
	DeclareString(n string, iv string) Builder
	DeclareFloat(n string, iv float64) Builder
	Build() (Built, error)
	MustBuild() Built
}

// This is the type produced by the Builder.Build() and Builder.MustBuild() methods
type Built interface {
	Open() Recorder
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
}

// This type is used by the View and the Registry to manage the
// life cycle and visibility of statistics within the registry
// and the view.
type Registration interface {
	Statistics
	// Increment the number observers
	Observe()
	// Decrement the number of observers.
	StopObserving() int
	// The number of open references to the receiver.
	Refs() int
}

// The type is used by the owner of the Statistics object to update it and
// to close it when done.
type Recorder interface {
	Statistics
	// Set a level statistics to a particular integer value
	SetInt(n string, i int64) Recorder
	// Set a level statistic to a particular float value
	SetFloat(n string, f float64) Recorder
	// Set a string statistic
	SetString(n string, s string) Recorder

	// Add an int value to an int statistic
	AddInt(n string, i int64) Recorder
	// Add a float value to a float statistic
	AddFloat(n string, f float64) Recorder
	// Drop one reference to the object obtained with Open(). T
	// The described object closes the Statistics object first, then the monitor.
	Close()
}

// The Registry type:
//   * allows described object to obtain a Builder
//   * allows the Built type to register an Recorder
//   * allows observers to obtain Views
type Registry interface {

	// Create a new Builder of statistics objects.
	NewBuilder(k string, n string, tags map[string]string) Builder

	// Called by a Statistics implementation to register itself when it
	// is first opened.
	Register(r Registration)

	// Open a view of all the registered statistics
	Open() View
}

// The View type provides a means for observers of the registry
// to view the contents of the registry.
//
// A View is guranteed to see IsOpen() == false on all Statistics
// that were open before or after the view was created and closed
// before the view is closed.
type View interface {
	// Called to iterate over the registered statistics sets.
	Do(f func(s Statistics)) View

	// Release all the observations held by a view.
	Close()
}
