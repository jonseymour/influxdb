package stats

import (
	"errors"
	"expvar"
	"strconv"
	"sync"
)

var ErrStatNotDeclared = errors.New("statistic has not been declared")
var ErrStatDeclaredWithDifferentType = errors.New("statistic declared with different type")

// This interface is used at runtime by objects that are described by Statistics
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

	OpenObserver() Statistics
	CloseObserver()

	// The number of open references to the receiver.
	Refs() int
}

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
	types      map[string]string
	built      bool
	refs       int
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

func (s *statistics) SetInt(n string, i int64) Owner {
	s.assertDeclaredAs(n, "int")
	s.intVars[n].Set(i)
	return s
}
func (s *statistics) SetFloat(n string, f float64) Owner {
	s.assertDeclaredAs(n, "float")
	s.floatVars[n].Set(f)
	return s
}

func (s *statistics) SetString(n string, v string) Owner {
	s.assertDeclaredAs(n, "string")
	s.stringVars[n].Set(v)
	return s
}

func (s *statistics) AddInt(n string, i int64) Owner {
	s.assertDeclaredAs(n, "int")
	s.intVars[n].Add(i)
	return s
}

func (s *statistics) AddFloat(n string, f float64) Owner {
	s.assertDeclaredAs(n, "float")
	s.floatVars[n].Add(f)
	return s
}

func (s *statistics) String() string {
	return s.impl.String()
}

// Consideration should be given to either commenting out the implementation
// or the calls to this method. In well-tested code, it will never do
// anything useful. The main reason for leaving it in is to document
// the requirement that the Statistics methods should never be called
// with a name which was not previously declared.
//
// One option might be to leave this method in during transition to the new statistics
// API to provide helpful error messages to developers who might not have grok'd the
// requirements of the new API properly, and then remove it once the code
// base has been transitioned.
//
// This will have the advantage of communicating the requirements of the new API
// to developers without imposing a long term cost on the runtime.
func (s *statistics) assertDeclaredAs(n string, t string) {
	if declared, ok := s.types[n]; !ok || t != declared {
		if !ok {
			panic(ErrStatNotDeclared)
		} else {
			panic(ErrStatDeclaredWithDifferentType)
		}
	}
}
