package stats

import (
	"errors"
	"expvar"
	"strconv"
	"sync"
)

var ErrAlreadyBuilt = errors.New("builder method must not be called in built state")

// This interface is used to declare the statistics fields during initialisation
// of a Statistics interface. A builder may be used at most once.
type Builder interface {
	DeclareInt(n string, iv int64) Builder
	DeclareString(n string, iv string) Builder
	DeclareFloat(n string, iv float64) Builder
	Build() (Openable, error)
	MustBuild() Openable
}

// This type forces the consumer of a Builder's Build() or MustBuild() method to issue an Open()
// call before attempting to use the StatistcsSet(). It helps to ensure that the
// location that invokes Open() is the owner of the responsibilty for issuing the subsequent
// Close()
type Openable interface {
	Open() Statistics
}

// This interface is used at runtime by objects that are described by Statistics
type Statistics interface {
	expvar.Var
	Openable

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

	// Set a level statistics to a particular integer value
	SetInt(n string, i int64) Statistics
	// Set a level statistic to a particular float value
	SetFloat(n string, f float64) Statistics
	// Set a string statistic
	SetString(n string, s string) Statistics

	// Add an int value to an int statistic
	AddInt(n string, i int64) Statistics
	// Add a float value to a float statistic
	AddFloat(n string, f float64) Statistics

	// The number of open references to the receiver.
	Refs() int
	// Drop one reference to the object obtained with Open(). T
	// The described object closes the Statistics object first, then the monitor.
	Close()
}

// The type which is used to implement both the Builder and Statistics interface
type statistics struct {
	mu         sync.RWMutex
	registry   Registry
	name       string
	key        string
	tags       map[string]string
	impl       *expvar.Map
	intVars    map[string]*expvar.Int
	stringVars map[string]*expvar.String
	floatVars  map[string]*expvar.Float
	built      bool
	refs       int
}

// Checks whether the receiver has already been built and returns an error if it has
func (s *statistics) checkNotBuilt() error {
	if s.built {
		return ErrAlreadyBuilt
	} else {
		return nil
	}
}

// Calls checkNotBuilt and panic if an error is returned
func (s *statistics) assertNotBuilt() {
	if err := s.checkNotBuilt(); err != nil {
		panic(err)
	}
}

// Declare an integer statistic
func (s *statistics) DeclareInt(n string, iv int64) Builder {
	s.assertNotBuilt()
	v := &expvar.Int{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.intVars[n] = v
	return s
}

// Declare a string statistic
func (s *statistics) DeclareString(n string, iv string) Builder {
	s.assertNotBuilt()
	v := &expvar.String{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.stringVars[n] = v
	return s
}

// Declare a float statistic
func (s *statistics) DeclareFloat(n string, iv float64) Builder {
	s.assertNotBuilt()
	v := &expvar.Float{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.floatVars[n] = v
	return s
}

// Finish building a Statistics returning an error on failure
func (s *statistics) Build() (Openable, error) {
	if err := s.checkNotBuilt(); err != nil {
		return nil, err
	}

	s.built = true
	tmp := &expvar.Map{}
	tmp.Init()
	s.impl.Do(func(kv expvar.KeyValue) {
		tmp.Set(kv.Key, kv.Value)
	})
	s.impl = tmp

	return s, nil
}

// Finish building a Statistics and panic on failure.
func (s *statistics) MustBuild() Openable {
	if set, err := s.Build(); err != nil {
		panic(err)
	} else {
		return set
	}
}

func (s *statistics) Key() string {
	return s.key
}

func (s *statistics) Name() string {
	return s.name
}

func (s *statistics) Tags() map[string]string {
	return s.tags
}

func (s *statistics) Map() *expvar.Map {
	return s.impl
}

func (s *statistics) Values() map[string]interface{} {
	values := make(map[string]interface{})
	n := s.Map()
	n.Do(func(kv expvar.KeyValue) {
		var f interface{}
		var err error
		switch v := kv.Value.(type) {
		case *expvar.Float:
			f, err = strconv.ParseFloat(v.String(), 64)
			if err != nil {
				return
			}
		case *expvar.Int:
			f, err = strconv.ParseInt(v.String(), 10, 64)
			if err != nil {
				return
			}
		default:
			f, err = strconv.Unquote(v.String())
			if err != nil {
				return
			}
		}
		values[kv.Key] = f
	})
	return values
}

func (s *statistics) SetInt(n string, i int64) Statistics {
	s.intVars[n].Set(i)
	return s
}
func (s *statistics) SetFloat(n string, f float64) Statistics {
	s.floatVars[n].Set(f)
	return s
}

func (s *statistics) SetString(n string, v string) Statistics {
	s.stringVars[n].Set(v)
	return s
}

func (s *statistics) AddInt(n string, i int64) Statistics {
	s.intVars[n].Add(i)
	return s
}

func (s *statistics) AddFloat(n string, f float64) Statistics {
	s.floatVars[n].Add(f)
	return s
}

func (s *statistics) String() string {
	return s.impl.String()
}

// Return true if there is less than 2 references to the receiver
func (s *statistics) Open() Statistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.refs++

	if s.refs == 2 {
		s.registry.Open(s)
	}
	return s
}

// Return true if there is less than 2 references to the receiver
func (s *statistics) Refs() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.refs
}

// Release one reference to the receiver.
func (s *statistics) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refs--
}
