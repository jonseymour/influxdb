// +build !race

package tsm1_test

import (
	"fmt"
	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

func TestCheckConcurrentReadsAreSafe(t *testing.T) {
	values := make(tsm1.Values, 1000)
	timestamps := make([]time.Time, len(values))
	series := make([]string, 100)
	for i := range timestamps {
		timestamps[i] = time.Unix(int64(rand.Int63n(int64(len(values)))), 0).UTC()
	}

	for i := range values {
		values[i] = tsm1.NewValue(timestamps[i*len(timestamps)/len(values)], float64(i))
	}

	for i := range series {
		series[i] = fmt.Sprintf("series%d", i)
	}

	wg := sync.WaitGroup{}
	c := tsm1.NewCache(1000000, "")

	ch := make(chan struct{})
	durations := make(chan time.Duration)
	for _, s := range series {
		for _, v := range values {
			c.Write(s, tsm1.Values{v})
		}
		reader := func(s string) {
			defer wg.Done()
			<-ch
			started := time.Now()
			c.Values(s)
			durations <- time.Now().Sub(started)
		}
		wg.Add(3)
		go reader(s)
		go reader(s)
		go reader(s)
	}

	nReaders := len(series) * 3
	var total time.Duration
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := nReaders
		for i > 0 {
			total += <-durations
			i--
		}

	}()
	close(ch)
	wg.Wait()
	fmt.Fprintf(os.Stderr, "average duration - %d (ns)\n", int64(total)/int64(nReaders))
}
