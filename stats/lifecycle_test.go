package stats_test

import (
	"reflect"
	"testing"

	"github.com/influxdata/influxdb/stats"
)

func TestNotifyOpenOrderStatisticFirst(t *testing.T) {
	stat := stats.Root.
		NewBuilder("key", "name", map[string]string{"tag": "T"}).
		MustBuild().
		Open()
	defer stat.Close()

	observed := []stats.Statistics{}
	defer func() {
		for _, e := range observed {
			e.Close()
		}
	}()

	closer := stats.Root.OnOpen(func(s stats.Openable) {
		observed = append(observed, s.Open())
	})
	defer closer()

	expected := []stats.Statistics{stat}
	if !reflect.DeepEqual(expected, observed) {
		t.Fatalf("did not observe existing statistic. got: %+v, expected: %+v", observed, expected)
	}
}

func TestNotifyOpenOrderObserverFirst(t *testing.T) {
	observed := []stats.Statistics{}
	defer func() {
		for _, e := range observed {
			e.Close()
		}
	}()

	closer := stats.Root.OnOpen(func(s stats.Openable) {
		observed = append(observed, s.Open())
	})
	defer closer()

	stat := stats.Root.
		NewBuilder("key", "name", map[string]string{"tag": "T"}).
		MustBuild().
		Open()
	defer stat.Close()

	expected := []stats.Statistics{stat}
	if !reflect.DeepEqual(expected, observed) {
		t.Fatalf("did not observe new statistic. got: %+v, expected: %+v", observed, expected)
	}
}
