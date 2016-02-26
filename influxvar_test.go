package influxdb_test

import (
	"expvar"
	"testing"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/stats"
)

func TestEmptyStatistics(t *testing.T) {
	found := make([]expvar.KeyValue, 0)
	influxdb.DoStatistics(func(kv expvar.KeyValue) {
		found = append(found, kv)
	})

	if length := len(found); length != 0 {
		t.Fatalf("non empty initial state. got %d, expected: %d", length, 0)
	}
}

// Test that we can create one statistic and that it disappears after it is deleted twice.
func TestOneStatistic(t *testing.T) {

	foo := influxdb.NewStatistics("foo", "bar", map[string]string{"tag": "T"})

	found := make([]expvar.KeyValue, 0)
	influxdb.DoStatistics(func(kv expvar.KeyValue) {
		found = append(found, kv)
	})

	if len(found) != 1 {
		t.Fatalf("enumeration error after do. length of slice: got %d, expected %d", len(found), 1)
	}
	if m, ok := found[0].Value.(stats.Statistics); !ok {
		t.Fatalf("value of found object got: %v, expected: a Statistics", found[0].Value)
	} else {
		if fooActual := m.Map(); fooActual != foo {
			t.Fatalf("failed to obtain expected map. got: %v, expected: %v", fooActual, foo)
		}
	}

	influxdb.CloseStatistics("foo")

	found = make([]expvar.KeyValue, 0)
	influxdb.DoStatistics(func(kv expvar.KeyValue) {
		found = append(found, kv)
		set := kv.Value.(stats.Statistics)
		if set.Refs() == 1 {
			set.Close()
		}
	})

	if length := len(found); length != 1 {
		t.Fatalf("failed to find expected number of objects. got: %d, expected: 1", length)
	}

	found = make([]expvar.KeyValue, 0)
	influxdb.DoStatistics(func(kv expvar.KeyValue) {
		found = append(found, kv)
	})

	if length := len(found); length != 0 {
		t.Fatalf("failed to find expected number of objects. got: %d, expected: 0", length)
	}
}
