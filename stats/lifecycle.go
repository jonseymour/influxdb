package stats

// Return true if there is less than 2 references to the receiver
func (s *statistics) Open() Statistics {
	var notify bool

	s.mu.RLock()
	s.refs++
	notify = (s.refs == 1)
	s.mu.RUnlock()

	// Perform this notification outside of a lock.
	// Inside of a lock, there is no room to move.
	//
	// With apologies to Groucho Marx.
	if notify {
		s.registry.NotifyOpen(s)
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

// Used by newly opened Statistics objects to notify OnOpen
// listeners that a new Statistics object has been registered.
func (r *registry) NotifyOpen(s Statistics) {

	// clone the list of listeners
	r.mu.RLock()
	clone := make([]*listener, len(r.listeners))
	copy(clone, r.listeners)
	r.mu.RUnlock()

	// call the each of the cloned listeners without holding any lock
	for _, l := range clone {
		l.callback(s)
	}

	// update the statistics map while holding the write lock
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getStatistics().Set(s.Key(), s)

	return
}

// Register a new OnOpen listener. The listener will receive notifications for
// all open Statistics currently in the Registry and for any objects that are
// subsequently added.
func (r *registry) OnOpen(lf func(o Openable)) func() {

	existing := []Statistics{}

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

	r.do(func(s Statistics) {
		existing = append(existing, s)
	})

	r.listeners = append(r.listeners, l)
	r.mu.Unlock()

	// Call the listener on objects that were already in the map before we added a listener.
	for _, s := range existing {
		if s.Refs() > 0 {
			lf(s)
		}
	}

	// By the time we get here, the listener has received one notification for
	// each Statistics object that was in the map prior to the listener being registered
	// and one notification for each added since. The notifications won't necessarily be received
	// in order of their original delivery to other listeners.

	return l.closer
}
