package stats

import (
	"errors"
	"expvar"

	"github.com/influxdata/influxdb"
)

var ErrAlreadyBuilt = errors.New("builder method must not be called in built state")

// This interface is used to declare the statistics fields during initialisation
// of a StatisticsSet interface. A builder may be used at most once.
type StatisticsSetBuilder interface {
	DeclareInt(n string, iv int64) StatisticsSetBuilder
	DeclareString(n string, iv string) StatisticsSetBuilder
	DeclareFloat(n string, iv float64) StatisticsSetBuilder
	Build() (StatisticsSet, error)
	MustBuild() StatisticsSet
}

// This interface is used at runtime by objects that are described by StatisticsSet
type StatisticsSet interface {
	// The statistic key
	Key() string
	// The statistic name
	Name() string
	// The statistic tags
	Tags() map[string]string
	// The underlying values map
	Map() *expvar.Map

	// Set a level statistics to a particular integer value
	SetInt(n string, i int64) StatisticsSet
	// Set a level statistic to a particular float value
	SetFloat(n string, f float64) StatisticsSet
	// Set a string statistic
	SetString(n string, s string) StatisticsSet

	AddInt(n string, i int64) StatisticsSet
	// Add a float value to a float statistic
	AddFloat(n string, f float64) StatisticsSet

	// Stop publishing statistics for this set.
	Close()
}

// The type which is used to implement both the StatisticsSetBuilder and StatisticsSet interface
type statistics struct {
	name       string
	key        string
	tags       map[string]string
	impl       *expvar.Map
	intVars    map[string]*expvar.Int
	stringVars map[string]*expvar.String
	floatVars  map[string]*expvar.Float
	built      bool
}

// A constructor for a new set of statistics
func NewStatisticsSetBuilder(k string, n string, tags map[string]string) StatisticsSetBuilder {
	impl := &expvar.Map{}
	impl.Init()
	return &statistics{
		key:        k,
		name:       n,
		tags:       tags,
		impl:       impl,
		intVars:    map[string]*expvar.Int{},
		stringVars: map[string]*expvar.String{},
		floatVars:  map[string]*expvar.Float{},
	}
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
func (s *statistics) DeclareInt(n string, iv int64) StatisticsSetBuilder {
	s.assertNotBuilt()
	v := &expvar.Int{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.intVars[n] = v
	return s
}

// Declare a string statistic
func (s *statistics) DeclareString(n string, iv string) StatisticsSetBuilder {
	s.assertNotBuilt()
	v := &expvar.String{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.stringVars[n] = v
	return s
}

// Declare a float statistic
func (s *statistics) DeclareFloat(n string, iv float64) StatisticsSetBuilder {
	s.assertNotBuilt()
	v := &expvar.Float{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.floatVars[n] = v
	return s
}

// Finish building a StatisticsSet returning an error on failure
func (s *statistics) Build() (StatisticsSet, error) {
	if err := s.checkNotBuilt(); err != nil {
		return nil, err
	}

	s.built = true

	tmp := influxdb.NewStatistics(s.key, s.name, s.tags)
	s.impl.Do(func(kv expvar.KeyValue) {
		tmp.Set(kv.Key, kv.Value)
	})
	s.impl = tmp

	return s, nil
}

// Finish building a StatisticsSet and panic on failure.
func (s *statistics) MustBuild() StatisticsSet {
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

func (s *statistics) SetInt(n string, i int64) StatisticsSet {
	s.intVars[n].Set(i)
	return s
}
func (s *statistics) SetFloat(n string, f float64) StatisticsSet {
	s.floatVars[n].Set(f)
	return s
}

func (s *statistics) SetString(n string, v string) StatisticsSet {
	s.stringVars[n].Set(v)
	return s
}

func (s *statistics) AddInt(n string, i int64) StatisticsSet {
	s.intVars[n].Add(i)
	return s
}

func (s *statistics) AddFloat(n string, f float64) StatisticsSet {
	s.floatVars[n].Add(f)
	return s
}

func (s *statistics) String() string {
	return s.impl.String()
}

func (s *statistics) Close() {
	influxdb.CloseStatistics(s.key)
}
