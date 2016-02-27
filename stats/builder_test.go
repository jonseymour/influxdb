package stats_test

import (
	"reflect"
	"testing"

	"github.com/influxdata/influxdb/stats"
)

type TestData struct {
	method string
	name   string
	value  interface{}
	err    error
}

func (d *TestData) applyToBuilder(t *testing.T, b stats.Builder) {
	result := func() (result error) {

		defer func() {
			e := recover()
			if e != nil {
				result = e.(error)
			}
		}()

		switch d.method {
		case "DeclareInt":
			b.DeclareInt(d.name, d.value.(int64))
		case "DeclareFloat":
			b.DeclareFloat(d.name, d.value.(float64))
		case "DeclareString":
			b.DeclareString(d.name, d.value.(string))
		default:
			t.Fatalf("unsupported method: %s", d.method)
		}

		return nil
	}()
	if d.err != result {
		t.Fatalf("unexpected result for method call: %s(%s, %v). got: %v, expected: %v", d.method, d.name, d.value, result, d.err)
	}
}

func (d *TestData) applyToStatistics(t *testing.T, s stats.Statistics) {
	result := func() (result error) {

		defer func() {
			e := recover()
			if e != nil {
				result = e.(error)
			}
		}()

		switch d.method {
		case "AddInt":
			s.AddInt(d.name, d.value.(int64))
		case "AddFloat":
			s.AddFloat(d.name, d.value.(float64))
		case "SetInt":
			s.SetInt(d.name, d.value.(int64))
		case "SetFloat":
			s.SetFloat(d.name, d.value.(float64))
		case "SetString":
			s.SetString(d.name, d.value.(string))
		default:
			t.Fatalf("unsupported method: %s", d.method)
		}

		return nil
	}()
	if d.err != result {
		t.Fatalf("unexpected result for method call: %s(%s, %v). got: %v, expected: %v", d.method, d.name, d.value, result, d.err)
	}
}

func TestDeclares(t *testing.T) {
	builder := stats.Root.
		NewBuilder("k", "n", map[string]string{"tags": "T"})

	for _, d := range []TestData{
		TestData{method: "DeclareInt", name: "intv", value: int64(1), err: nil},
		TestData{method: "DeclareFloat", name: "floatv", value: float64(1.5), err: nil},
		TestData{method: "DeclareString", name: "stringv", value: "1", err: nil},
		TestData{method: "DeclareInt", name: "intv", value: int64(2), err: stats.ErrStatAlreadyDeclared},
		TestData{method: "DeclareFloat", name: "floatv", value: float64(2.0), err: stats.ErrStatAlreadyDeclared},
		TestData{method: "DeclareString", name: "stringv", value: "2", err: stats.ErrStatAlreadyDeclared},
	} {
		d.applyToBuilder(t, builder)
	}

	open := builder.MustBuild().Open()
	values := open.Values()
	open.Close()

	expected := map[string]interface{}{
		"stringv": "1",
		"floatv":  1.5,
		"intv":    int64(1),
	}

	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("incorrect values found. got: %v, expected: %v", values, expected)
	}
}

func TestMustBuild(t *testing.T) {
	builder := stats.Root.
		NewBuilder("k", "n", map[string]string{"tags": "T"})

	for _, d := range []TestData{
		TestData{method: "DeclareInt", name: "intv", value: int64(1), err: nil},
		TestData{method: "DeclareFloat", name: "floatv", value: float64(1.5), err: nil},
		TestData{method: "DeclareString", name: "stringv", value: "1", err: nil},
	} {
		d.applyToBuilder(t, builder)
	}

	open := builder.MustBuild().Open()
	expected := map[string]interface{}{
		"stringv": "1",
		"floatv":  1.5,
		"intv":    int64(1),
	}
	values := open.Values()

	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("incorrect values found. got: %v, expected: %v", values, expected)
	}

	for _, d := range []TestData{
		TestData{method: "SetInt", name: "intv", value: int64(2), err: nil},
		TestData{method: "SetFloat", name: "floatv", value: float64(2.5), err: nil},
		TestData{method: "SetString", name: "stringv", value: "2", err: nil},
	} {
		d.applyToStatistics(t, open)
	}

	expected = map[string]interface{}{
		"stringv": "2",
		"floatv":  2.5,
		"intv":    int64(2),
	}
	values = open.Values()

	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("incorrect values found. got: %v, expected: %v", values, expected)
	}

	for _, d := range []TestData{
		TestData{method: "AddInt", name: "intv", value: int64(3), err: nil},
		TestData{method: "AddFloat", name: "floatv", value: float64(3.5), err: nil},
	} {
		d.applyToStatistics(t, open)
	}

	expected = map[string]interface{}{
		"stringv": "2",
		"floatv":  6.0,
		"intv":    int64(5),
	}
	values = open.Values()

	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("incorrect values found. got: %v, expected: %v", values, expected)
	}

	for _, d := range []TestData{
		TestData{method: "SetFloat", name: "intv", value: float64(4.0), err: stats.ErrStatDeclaredWithDifferentType},
		TestData{method: "SetString", name: "floatv", value: "4.6", err: stats.ErrStatDeclaredWithDifferentType},
		TestData{method: "SetInt", name: "stringv", value: int64(4), err: stats.ErrStatDeclaredWithDifferentType},
	} {
		d.applyToStatistics(t, open)
	}

	open.Close()

}
