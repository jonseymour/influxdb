package stats

import (
	"errors"
)

// The error thrown if there is an expected reference counting violation
var errUnexpectedRefCount = errors.New("unexpected reference counting error")

// Open the Recorder and register it with the registryClient
func (s *statistics) Open() Recorder {
	s.open(true)
	return s
}

// Close the Recorder.
func (s *statistics) Close() {
	s.close(true)
}

// Increment the reference count,
// set the isOpen() status and conditionally notify the
// registry of the new Recorder
func (s *statistics) open(owner bool) {
	s.mu.Lock()
	if owner {
		if s.isRecorderOpen {
			s.mu.Unlock()
			panic(ErrAlreadyOpen)
		}
		s.isRecorderOpen = true
	}
	s.refsCount++
	s.mu.Unlock()

	// Perform this notification outside of a lock.
	// Inside of a lock, there is no room to move.
	//
	// With apologies to Groucho Marx.
	if owner {
		s.registry.register(s)
	}
}

// Decrement the reference count and conditionally
// clear the isOpen().
func (s *statistics) close(owner bool) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.refsCount == 0 {
		panic(errUnexpectedRefCount)
	}

	if owner {
		if !s.isRecorderOpen {
			panic(ErrAlreadyClosed)
		}
		s.isRecorderOpen = false
	}
	s.refsCount--
	return s.refsCount
}

// True if the Recorder interface is still open.
func (s *statistics) isOpen() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRecorderOpen
}

// Return true if there is less than 2 references to the receiver
func (s *statistics) refs() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.refsCount
}

// Register an observer.
func (s *statistics) observe() {
	s.open(false)
}

// Deregister an observer.
func (s *statistics) stopObserving() int {
	return s.close(false)
}

// Used by newly opened Statistics objects to notify OnOpen
// listeners that a new Statistics object has been registered.
func (r *registry) register(g registration) {

	// clone the list of listeners
	r.mu.RLock()
	clone := make([]*listener, len(r.listeners))
	copy(clone, r.listeners)
	r.mu.RUnlock()

	// call the each of the cloned listeners without holding any lock
	for _, l := range clone {
		l.callback(g)
	}

	// update the statistics map while holding the write lock
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getStatistics().Set(g.Key(), g)

	return
}

// Register a new OnOpen listener. The listener will receive notifications for
// all open Statistics currently in the Registry and for any objects that are
// subsequently added.
func (r *registry) onOpen(lf func(o registration)) func() {

	existing := []registration{}

	// add a new listener while holding the write lock
	r.mu.Lock()
	l := &listener{
		callback: lf,
	}
	l.closer = func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		for i, e := range r.listeners {
			if e == l {
				r.listeners = append(r.listeners[:i], r.listeners[i+1:]...)
				return
			}
		}
	}

	r.do(func(s registration) {
		existing = append(existing, s)
	})

	r.listeners = append(r.listeners, l)
	r.mu.Unlock()

	// Call the listener on objects that were already in the map before we added a listener.
	for _, g := range existing {
		if g.isOpen() {
			lf(g)
		}
	}

	// By the time we get here, the listener has received one notification for
	// each Statistics object that was in the map prior to the listener being registered
	// and one notification for each added since. The notifications won't necessarily be received
	// in order of their original delivery to other listeners.

	return l.closer
}
