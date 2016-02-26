package stats_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/influxdb/stats"
)

// Tests that there are no deadlocks or races under typical usage scenarios
func TestLifeCycleRaces(t *testing.T) {

	errors := make(chan error)
	go func() {
		defer func() {
			e := recover()
			if e != nil {
				errors <- fmt.Errorf("panic: %v", e)
			} else {
				errors <- nil
			}
		}()

		nwriters := 10
		nloops := 5
		start := make(chan struct{})
		stop := make(chan struct{})
		wg := sync.WaitGroup{}

		seen := map[string]struct{}{}

		monitor := func() {
			defer wg.Done()
			<-start
			closer := stats.Root.OnOpen(func(s stats.Openable) {
				s.Open()
			})
			defer closer()

			iterator := func(s stats.Statistics) {
				seen[s.Key()] = struct{}{}
				s.Values()
				if s.Refs() == 1 {
					s.Close()
				}
			}

			for {
				select {
				case _ = <-stop:
					stats.Root.Do(iterator)
					return
				default:
					stats.Root.Do(iterator)
				}
			}
		}

		writer := func(i int) {
			defer wg.Done()
			<-start

			k := fmt.Sprintf("stat:%d", i)
			stats := stats.Root.
				NewBuilder(k, k, map[string]string{"tag": "T"}).
				DeclareInt("value", 0).
				DeclareInt("index", 0).
				MustBuild().
				Open()

			defer stats.Close()

			for stats.Refs() < 2 {
				for n := 0; n < nloops; n++ {
					stats.
						AddInt("value", int64(1)).
						SetInt("index", int64(n))
					time.Sleep(time.Microsecond * 100)
				}
			}
		}

		for i := 0; i < nwriters; i++ {
			wg.Add(1)
			go writer(i)

			// start the monitor after some goroutines started
			if i == nwriters/2 {
				go monitor()
			}
		}

		close(start)
		wg.Wait() // wait until all the writers stop

		wg.Add(1)
		close(stop)
		wg.Wait() // wait for the monitor to stop

		// check that there are no statistics still registered
		count := 0
		stats.Root.Do(func(s stats.Statistics) {
			count++
		})
		if count != 0 {
			t.Fatalf("too many registered statistics. got: %d, expected: 0", count)
		}

		if len(seen) != nwriters {
			t.Fatalf("failed to observe some statistics. got: %d, expected: %d", len(seen), nwriters)
		}
	}()
	select {
	case err := <-errors:
		if err != nil {
			t.Fatalf("got: %v, expected: nil", err)
		}
	case <-time.NewTimer(time.Second * 5).C:
		// force a stack dump to allow analysis of issue
		panic("got: timeout, expected: normal return")
	}
}
