// The types defined in this module reflect different states or roles of a Statistics objects.
//
// Consumers of Statistics objects get exactly the interface they need to perform
// the task their role requires and no more.
//
// For example, consider the life cycle of a statistics object.
//
// It starts out as a Builder, until it is built when it becomes Built. To obtain
// access to the Recorder interface, the owner must execute Built.Open() which both yields
// the Recorder instance and registers that instance with the Registry.  The Recorer interface
// provides access to the Set*() and Add*() methods and a Close() method
// to be called at the end of the described object's life cycle.
//
package stats

import (
	"expvar"
)

// This interface is used to declare the statistics fields during initialisation
// of a Recorder instance.
//
// Any given name may be used as the parameter of a Declare*() method at most once.
//
// It is an error to call the Declare*()
// methods after Build() or MustBuild() have been called.
//
// The Build() and MustBuild() methods of Builder may be used at most once.
//
// To obtain a Builder instance, call stats.Root.NewBuilder().
type Builder interface {

	// The Declare methods are used to declare which statistics the resulting
	// Recorder will write.

	// Declare an integer statistic.
	DeclareInt(n string, iv int64) Builder
	// Declare an string statistic
	DeclareString(n string, iv string) Builder
	// Declare an float statistic
	DeclareFloat(n string, iv float64) Builder

	// Build a Built instance or return an error.
	Build() (Built, error)
	// Build a Built instance or panic.
	MustBuild() Built
}

// This is the type produced by the Builder.Build() and Builder.MustBuild() methods. To
// simultaneously obtain a reference to the Recorder and register that Recorder instance
// with the related Regsitry call the Open() method.
//
// To obtain a Built instance call the Build() or MustBuild() methods of a Builder instance.
type Built interface {
	Open() Recorder
}

// This type is used by observers to read the state of a single Statistics object.
type Statistics interface {
	expvar.Var

	// The statistic key
	Key() string
	// The statistic name
	Name() string
	// The statistic tags
	Tags() map[string]string

	// A map of the statistics using their corrsponding native go types.
	Values() map[string]interface{}

	// The underlying values map - do not use directly, provided for legacy expvar support only.
	Map() *expvar.Map
}

// The Recorder type is used by described objects to update statistics either by
// setting new values (for level-type statistics) or adding to an existing value (counter-type)
// statistics. Only statistic names which were declared during the construction of the Recorder
// may be used with a Recorder's Set and Add methods.
//
// At the end of the described object's life cycle, the described object should call Recorder.Close() method.
// This will release the resources associated with the Recorder and prevent any observers from publishing
// additional updates for that Recorder.
//
// Remember to close the Recorder instance at the end of the described object's life cycle by calling
// the Recorder.Close() method.
//
// To obtain a Recorder, call the Open() method of a Built instance. For example:
//
//     var recorder Recorder = stats.Root.
//        NewBuilder("key", "name", map[string]string{"tag": "T"}).
//        DeclareInt("counter",0).
//        MustBuild().
//        Open()
//
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

// The Registry type allows described objects to obtain a Builders used
// to construct their Recorder instances. For example: a described object might be
// implemented like this:
//
//    import "github.com/influxdata/influxdb/stats"
//
//    type FooBar struct {
//        stats stats.Recorder
//    }
//
//    ...
//
//    fooBar := &FooBar{
//        stats: stats.Root.NewBuilder("foobar:1", "foobar:1", map[string]string{"tags": "T"}).
//           DeclareInt("counter", 0).
//           MustBuild().
//           Open(),
//    }
//
//    ...
//
//    func (fb * FooBar) Close() {
//        fb.stats.Close()
//    }
//
// The Registry type also allows observers to obtain View instances which
// allows iteration over the contents of the registry.
//
//   view := stats.Root.Open()
//   defer view.Close()
//
//   view.Do(
//      func(s Statistics) {
//         // do something with a Statistics object
//         ...
//      }
//   )
//
type Registry interface {

	// Create a new Builder of statistics objects.
	NewBuilder(k string, n string, tags map[string]string) Builder

	// Open a view of all the registered statistics
	Open() View
}

// The View type provides a means for observers of the registry
// to view the contents of the registry.
//
// Views are obtained by calling stats.Root.Open().
type View interface {
	// Called to iterate over the registered statistics sets.
	Do(f func(s Statistics)) View

	// Release all the observations held by a view.
	Close()
}

var (
	// The reference to the Registry singleton.
	Root Registry

	// The panic that is issued if an attempt is made to record an undeclared statistic.
	ErrStatNotDeclared = errors.New("statistic has not been declared")

	// The panic that is issued if an attempt is made record a statistic declared with a different type.
	ErrStatDeclaredWithDifferentType = errors.New("statistic declared with different type")

	// The panic that is issued if the Builder.Build() or Builder.MustBuild() method is called more than once on the same Builder.
	ErrAlreadyBuilt = errors.New("builder method must not be called in built state")

	// The panic that is issued if the same statistic is declared twice.
	ErrStatAlreadyDeclared = errors.New("statistic has already been declared")
)

func init() {
	Root = &registry{
		listeners: make([]*listener, 0),
	}
}
