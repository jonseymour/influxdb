package stats

import (
	"errors"
	"expvar"
	"strconv"
	"sync"
)

var ErrStatNotDeclared = errors.New("statistic has not been declared")
var ErrStatDeclaredWithDifferentType = errors.New("statistic declared with different type")

// The type which is used to implement both the Builder and Statistics interface
type statistics struct {
	mu         sync.RWMutex
	registry   registryClient
	name       string
	key        string
	tags       map[string]string
	impl       *expvar.Map
	intVars    map[string]*expvar.Int
	stringVars map[string]*expvar.String
	floatVars  map[string]*expvar.Float
	types      map[string]string
	built      bool
	isOpen     bool
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

func (s *statistics) SetInt(n string, i int64) Recorder {
	s.assertDeclaredAs(n, "int")
	s.intVars[n].Set(i)
	return s
}
func (s *statistics) SetFloat(n string, f float64) Recorder {
	s.assertDeclaredAs(n, "float")
	s.floatVars[n].Set(f)
	return s
}

func (s *statistics) SetString(n string, v string) Recorder {
	s.assertDeclaredAs(n, "string")
	s.stringVars[n].Set(v)
	return s
}

func (s *statistics) AddInt(n string, i int64) Recorder {
	s.assertDeclaredAs(n, "int")
	s.intVars[n].Add(i)
	return s
}

func (s *statistics) AddFloat(n string, f float64) Recorder {
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
