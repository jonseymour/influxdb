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
	collector := func(s stats.Statistics) {
		observed = append(observed, s)
	}
	stats.Root.Open().Do(collector).Close()

	expected := []stats.Statistics{stat}
	if !reflect.DeepEqual(expected, observed) {
		t.Fatalf("did not observe existing statistic. got: %+v, expected: %+v", observed, expected)
	}
}

func TestNotifyOpenOrderObserverFirst(t *testing.T) {
	observed := []stats.Statistics{}
	collector := func(s stats.Statistics) {
		observed = append(observed, s)
	}

	view := stats.Root.Open()
	defer view.Close()

	stat := stats.Root.
		NewBuilder("key", "name", map[string]string{"tag": "T"}).
		MustBuild().
		Open()
	defer stat.Close()

	view.Do(collector)

	expected := []stats.Statistics{stat}
	if !reflect.DeepEqual(expected, observed) {
		t.Fatalf("did not observe new statistic. got: %+v, expected: %+v", observed, expected)
	}
}
