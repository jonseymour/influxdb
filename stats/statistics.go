package stats

import (
	"expvar"
	"strconv"
	"sync"
)

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
