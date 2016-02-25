package influxdb_test

import (
	"expvar"
	"testing"

	"github.com/influxdata/influxdb"
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
	if m, ok := found[0].Value.(*expvar.Map); !ok {
		t.Fatalf("value of found object got: %v, expected: a map", found[0].Value)
	} else {
		if fooActual := m.Get("values"); fooActual != foo {
			t.Fatalf("failed to obtain expected map. got: %v, expected: %v", fooActual, foo)
		}
	}

	influxdb.DeleteStatistics("foo")

	found = make([]expvar.KeyValue, 0)
	influxdb.DoStatistics(func(kv expvar.KeyValue) {
		found = append(found, kv)
	})

	if length := len(found); length != 1 {
		t.Fatalf("failed to find expected number of objects. got: %d, expected: 1", length)
	}

	influxdb.DeleteStatistics("foo")

	found = make([]expvar.KeyValue, 0)
	influxdb.DoStatistics(func(kv expvar.KeyValue) {
		found = append(found, kv)
	})

	if length := len(found); length != 0 {
		t.Fatalf("failed to find expected number of objects. got: %d, expected: 0", length)
	}
}
