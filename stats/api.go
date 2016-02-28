// The 'stats' package defines a statistics acquisition and monitoring API
// which uses the go 'expvar' API as the underlying representation mechanism.
//
// Their are several advantages of using this API to manage the expvar namespace
// over the raw expvar API. The main one is that it trivially allows items to
// be removed from the expvar namespace in a way that allows viewers of the namespace
// to see the final update to a expvar Map immediately prior to the removal of the Map
// from the namespace. This is achieved by use of the View type which encapsulates
// a reference counting mechanism which ensures a Statistics object does not disappear from a
// View until the View has seen the last state of the object prior to it being
// closed by the owner of its Recorder.
//
// The ability to easily remove maps from the 'expvar' namespace is important in long-lived
// processes that create large number of short-lived variables in the 'expvar' namespace - unless
// some cleanup mechanism such as the one provided by this package is used resources
// of various kind (memory, CPU or IO) might be exhausted.
//
// The API is designed so that consumers of Statistics objects get exactly
// the interface they need to perform the task their role requires and no more.
//
// For example, consider the life cycle of a Statistics object.
//
// It starts out as a Builder, obtained from stats.Root.NewBuilder(), by the object
// it describes (the so-called 'described object'). The described object then configures
// the Builder by declaring which statistics will be written into the Recorder when the
// Recorder becomes live. Once the declarations are finished, the described object calls
// MustBuild() to obtain a reference to a frozen interface of type Built. Assuming this
// call succeeds, the caller then invokes the Built.Open() method to simultaneously
// register and obtain a reference to the Recorder interface that will be used at runtime.
// The Recorder interface exposes Set and Add methods that allow lock-free access
// to slices of the underlying expvar Map variables and a Close() method that can be used
// by the described object to release resources associated with the Recorder and remove
// the Statistics object from any open Views.
//
package stats

import (
	"errors"
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
	// The panic that is issued if the same statistic is declared twice.
	ErrStatAlreadyDeclared = errors.New("statistic has already been declared")

	// The panic that is issued if the Builder.Build() or Builder.MustBuild() method is called more than once on the same Builder.
	ErrAlreadyBuilt = errors.New("builder method must not be called in built state")

	// The panic that is issued if an attempt is made to record an undeclared statistic.
	ErrStatNotDeclared = errors.New("statistic has not been declared")

	// The panic that is issued if an attempt is made record a statistic declared with a different type.
	ErrStatDeclaredWithDifferentType = errors.New("statistic declared with different type")
)

// A Collection is a collection of statistics
type Collection []Statistics
