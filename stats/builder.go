package stats

import (
	"expvar"
)

// Initializes a new Builder and associates with the specified registry.
func newBuilder(k string, tags map[string]string, r registryClient) Builder {
	impl := &expvar.Map{}
	impl.Init()

	builder := &statistics{
		registry:   r,
		key:        k,
		tags:       tags,
		impl:       impl,
		refsCount:  0,
		intVars:    map[string]*expvar.Int{},
		stringVars: map[string]*expvar.String{},
		floatVars:  map[string]*expvar.Float{},
		types:      map[string]string{},
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

// Asserts that the specifed statistic has not already been declared.
func (s *statistics) assertNotDeclared(n string) {
	if _, ok := s.types[n]; ok {
		panic(ErrStatAlreadyDeclared)
	}
}

// Declare an integer statistic
func (s *statistics) DeclareInt(n string, iv int64) Builder {
	s.assertNotBuilt()
	s.assertNotDeclared(n)
	v := &expvar.Int{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.intVars[n] = v
	s.types[n] = "int"
	return s
}

// Declare a string statistic
func (s *statistics) DeclareString(n string, iv string) Builder {
	s.assertNotBuilt()
	s.assertNotDeclared(n)
	v := &expvar.String{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.stringVars[n] = v
	s.types[n] = "string"
	return s
}

// Declare a float statistic
func (s *statistics) DeclareFloat(n string, iv float64) Builder {
	s.assertNotBuilt()
	s.assertNotDeclared(n)
	v := &expvar.Float{}
	v.Set(iv)
	s.impl.Set(n, v)
	s.floatVars[n] = v
	s.types[n] = "float"
	return s
}

// Finish building a Statistics returning an error on failure
func (s *statistics) Build() (Built, error) {
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
func (s *statistics) MustBuild() Built {
	if set, err := s.Build(); err != nil {
		panic(err)
	} else {
		return set
	}
}
