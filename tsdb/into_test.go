package tsdb_test

import (
	"github.com/influxdb/influxdb/tsdb"
	"reflect"
	"testing"
)

func TestEnsureUniqueNonEmptyNamesEmptySlice(t *testing.T) {
	input := []string{}
	expected := []string{}

	if got := tsdb.EnsureUniqueNonEmptyNames(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("mismatched result. got: %v, expected: %v", got, expected)
	}
}

func TestEnsureUniqueNonEmptyNamesUniqueNames(t *testing.T) {
	input := []string{"foo", "bar"}
	expected := input

	if got := tsdb.EnsureUniqueNonEmptyNames(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("mismatched result. got: %v, expected: %v", got, expected)
	}
}

func TestEnsureUniqueNonEmptyNamesDuplicateNames(t *testing.T) {
	input := []string{"foo", "foo"}
	expected := []string{"foo", "foo_1"}

	if got := tsdb.EnsureUniqueNonEmptyNames(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("mismatched result. got: %v, expected: %v", got, expected)
	}
}

func TestEnsureUniqueNonEmptyNamesEmptyName(t *testing.T) {
	input := []string{"foo", ""}
	expected := []string{"foo", "_1"}

	if got := tsdb.EnsureUniqueNonEmptyNames(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("mismatched result. got: %v, expected: %v", got, expected)
	}
}

func TestEnsureUniqueNonEmptyNamesTwoEmptyNames(t *testing.T) {
	input := []string{"foo", "", ""}
	expected := []string{"foo", "_1", "_2"}

	if got := tsdb.EnsureUniqueNonEmptyNames(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("mismatched result. got: %v, expected: %v", got, expected)
	}
}

func TestEnsureUniqueNonEmptyNamesAlreadyUsed(t *testing.T) {
	input := []string{"foo", "_2", ""}
	expected := []string{"foo", "_2", "_2_2"}

	if got := tsdb.EnsureUniqueNonEmptyNames(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("mismatched result. got: %v, expected: %v", got, expected)
	}
}

func TestMemoisedEnsureUniqueNonEmptyNamesAlreadyUsed(t *testing.T) {
	mapper := tsdb.MemoisedEnsureUniqueNonEmptyNames()
	input := []string{"foo", "_2", ""}
	inputAgain := []string{"foo", "_2", ""}
	input2 := []string{"foo", "bar", "qux"}
	expected := []string{"foo", "_2", "_2_2"}
	expected2 := []string{"foo", "bar", "qux"}
	got := []string{}

	if got = mapper(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("mismatched result. got: %v, expected: %v", got, expected)
	}

	// test that we get an identical slice if the input has the same elements
	if gotAgain := mapper(inputAgain); !reflect.DeepEqual(gotAgain, expected) {
		t.Fatalf("mismatched result. got: %v, expected: %v", gotAgain, expected)
	} else {
		if &got[0] != &gotAgain[0] {
			t.Fatalf("mismatched result. got: different slice, expected: same slice")
		}
	}

	// test that if the input changes, we get a new result
	if got := mapper(input2); !reflect.DeepEqual(got, expected2) {
		t.Fatalf("mismatched result. got: %v, expected: %v", got, expected2)
	}
}
