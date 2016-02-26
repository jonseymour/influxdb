package stats_test

import (
	"testing"

	"github.com/influxdata/influxdb/stats"
)

func TestEmptyStatistics(t *testing.T) {
	found := make([]stats.Statistics, 0)
	collector := func(s stats.Statistics) {
		found = append(found, s)
	}
	stats.Root.Open().Do(collector).Close()

	if length := len(found); length != 0 {
		t.Fatalf("non empty initial state. got %d, expected: %d", length, 0)
	}
}

// Test that we can create one statistic and that it disappears after it is deleted twice.
func TestOneStatistic(t *testing.T) {

	found := make([]stats.Statistics, 0)
	collector := func(s stats.Statistics) {
		found = append(found, s)
	}

	foo := stats.Root.
		NewBuilder("foo", "bar", map[string]string{"tag": "T"}).
		MustBuild().
		Open()
	defer func() {
		if foo != nil {
			foo.Close()
		}
	}()

	stats.Root.Open().Do(collector).Close()

	if len(found) != 1 {
		t.Fatalf("enumeration error after do. length of slice: got %d, expected %d", len(found), 1)
	}
	m := found[0]
	if fooActual := m; fooActual != foo {
		t.Fatalf("failed to obtain expected map. got: %v, expected: %v", fooActual, foo)
	}

	foo.Close()
	foo = nil

	found = make([]stats.Statistics, 0)
	stats.Root.Open().Do(collector).Close()

	if length := len(found); length != 0 {
		t.Fatalf("failed to find expected number of objects. got: %d, expected: 0", length)
	}
}
