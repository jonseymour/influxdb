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
