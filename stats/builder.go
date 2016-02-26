package stats

import (
	"errors"
	"expvar"
)

var ErrAlreadyBuilt = errors.New("builder method must not be called in built state")

// This type forces the consumer of a Builder's Build() or MustBuild() method to issue an Open()
// call before attempting to use the StatistcsSet(). It helps to ensure that the
// location that invokes Open() is the owner of the responsibilty for issuing the subsequent
// Close()
type Openable interface {
	Open() Statistics
}

// This interface is used to declare the statistics fields during initialisation
// of a Statistics interface. A builder may be used at most once.
type Builder interface {
	DeclareInt(n string, iv int64) Builder
	DeclareString(n string, iv string) Builder
	DeclareFloat(n string, iv float64) Builder
	Build() (Openable, error)
	MustBuild() Openable
}

func newBuilder(k string, n string, tags map[string]string, r *registry) Builder {
	impl := &expvar.Map{}
	impl.Init()

	builder := &statistics{
		registry:   r,
		key:        k,
		name:       n,
		tags:       tags,
		impl:       impl,
		refs:       0,
		intVars:    map[string]*expvar.Int{},
		stringVars: map[string]*expvar.String{},
		floatVars:  map[string]*expvar.Float{},
	}

	return builder
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
