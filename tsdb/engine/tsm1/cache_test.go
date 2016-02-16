package tsm1

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang/snappy"
)

func TestCache_NewCache(t *testing.T) {
	c := NewCache(100)
	if c == nil {
		t.Fatalf("failed to create new cache")
	}

	if c.MaxSize() != 100 {
		t.Fatalf("new cache max size not correct")
	}
	if c.Size() != 0 {
		t.Fatalf("new cache size not correct")
	}
	if len(c.Keys()) != 0 {
		t.Fatalf("new cache keys not correct: %v", c.Keys())
	}
}

func TestCache_CacheWrite(t *testing.T) {
	v0 := NewValue(time.Unix(1, 0).UTC(), 1.0)
	v1 := NewValue(time.Unix(2, 0).UTC(), 2.0)
	v2 := NewValue(time.Unix(3, 0).UTC(), 3.0)
	values := Values{v0, v1, v2}
	valuesSize := uint64(v0.Size() + v1.Size() + v2.Size())

	c := NewCache(3 * valuesSize)

	if err := c.Write("foo", values); err != nil {
		t.Fatalf("failed to write key foo to cache: %s", err.Error())
	}
	if err := c.Write("bar", values); err != nil {
		t.Fatalf("failed to write key foo to cache: %s", err.Error())
	}
	if n := c.Size(); n != 2*valuesSize {
		t.Fatalf("cache size incorrect after 2 writes, exp %d, got %d", 2*valuesSize, n)
	}

	if exp, keys := []string{"bar", "foo"}, c.Keys(); !reflect.DeepEqual(keys, exp) {
		t.Fatalf("cache keys incorrect after 2 writes, exp %v, got %v", exp, keys)
	}
}

func TestCache_CacheWriteMulti(t *testing.T) {
	v0 := NewValue(time.Unix(1, 0).UTC(), 1.0)
	v1 := NewValue(time.Unix(2, 0).UTC(), 2.0)
	v2 := NewValue(time.Unix(3, 0).UTC(), 3.0)
	values := Values{v0, v1, v2}
	valuesSize := uint64(v0.Size() + v1.Size() + v2.Size())

	c := NewCache(3 * valuesSize)

	if err := c.WriteMulti(map[string][]Value{"foo": values, "bar": values}); err != nil {
		t.Fatalf("failed to write key foo to cache: %s", err.Error())
	}
	if n := c.Size(); n != 2*valuesSize {
		t.Fatalf("cache size incorrect after 2 writes, exp %d, got %d", 2*valuesSize, n)
	}

	if exp, keys := []string{"bar", "foo"}, c.Keys(); !reflect.DeepEqual(keys, exp) {
		t.Fatalf("cache keys incorrect after 2 writes, exp %v, got %v", exp, keys)
	}
}

func TestCache_CacheValues(t *testing.T) {
	v0 := NewValue(time.Unix(1, 0).UTC(), 0.0)
	v1 := NewValue(time.Unix(2, 0).UTC(), 2.0)
	v2 := NewValue(time.Unix(3, 0).UTC(), 3.0)
	v3 := NewValue(time.Unix(1, 0).UTC(), 1.0)
	v4 := NewValue(time.Unix(4, 0).UTC(), 4.0)

	c := NewCache(512)
	if deduped := c.Values("no such key"); deduped != nil {
		t.Fatalf("Values returned for no such key")
	}

	if err := c.Write("foo", Values{v0, v1, v2, v3}); err != nil {
		t.Fatalf("failed to write 3 values, key foo to cache: %s", err.Error())
	}
	if err := c.Write("foo", Values{v4}); err != nil {
		t.Fatalf("failed to write 1 value, key foo to cache: %s", err.Error())
	}

	expAscValues := Values{v3, v1, v2, v4}
	if deduped := c.Values("foo"); !reflect.DeepEqual(expAscValues, deduped) {
		t.Fatalf("deduped ascending values for foo incorrect, exp: %v, got %v", expAscValues, deduped)
	}
}

func TestCache_CacheSnapshot(t *testing.T) {
	segment1 := "segment1.wal"
	v0 := NewValue(time.Unix(2, 0).UTC(), 0.0)
	v1 := NewValue(time.Unix(3, 0).UTC(), 2.0)
	v2 := NewValue(time.Unix(4, 0).UTC(), 3.0)
	v3 := NewValue(time.Unix(5, 0).UTC(), 4.0)
	v4 := NewValue(time.Unix(6, 0).UTC(), 5.0)
	v5 := NewValue(time.Unix(1, 0).UTC(), 5.0)

	c := NewCache(512)
	if err := c.Write("foo", Values{v0, v1, v2, v3}); err != nil {
		t.Fatalf("failed to write 3 values, key foo to cache: %s", err.Error())
	}

	// Grab snapshot, and ensure it's as expected.
	snapshots := c.PrepareSnapshots([]string{segment1})
	if length := len(snapshots); length != 1 {
		t.Fatalf("invalid snapshots length, exp %d, got %d", 1, length)
	}

	// check that the snapshot we got is not nil
	snapshot := snapshots[0]
	if snapshot == nil {
		t.Fatalf("snapshot is nil")
	}

	// check that the snapshot has captured the segment files
	expFiles := [][]string{[]string{segment1}}
	if files := [][]string{snapshot.Files()}; !reflect.DeepEqual(expFiles, files) {
		t.Fatalf("files is incorrect, exp: %v, got %v", expFiles, files)
	}

	expValues := Values{v0, v1, v2, v3}
	if deduped := snapshot.values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("snapshotted values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}

	// Ensure cache is still as expected.
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("post-snapshot values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}

	// Write a new value to the cache.
	if err := c.Write("foo", Values{v4}); err != nil {
		t.Fatalf("failed to write post-snap value, key foo to cache: %s", err.Error())
	}
	expValues = Values{v0, v1, v2, v3, v4}
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("post-snapshot write values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}

	// Write a new, out-of-order, value to the cache.
	if err := c.Write("foo", Values{v5}); err != nil {
		t.Fatalf("failed to write post-snap value, key foo to cache: %s", err.Error())
	}
	expValues = Values{v5, v0, v1, v2, v3, v4}
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("post-snapshot out-of-order write values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}

	// Clear snapshot, ensuring non-snapshot data untouched.
	c.CommitSnapshots()

	expValues = Values{v5, v4}
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("post-commit values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}
}

func TestCache_CacheRollbackSnapshots(t *testing.T) {
	segment2 := "segment2.wal"
	segment3 := "segment3.wal"
	segment4 := "segment4.wal"
	segment5 := "segment5.wal"
	v4 := NewValue(time.Unix(6, 0).UTC(), 5.0)
	v5 := NewValue(time.Unix(1, 0).UTC(), 5.0)
	v6 := NewValue(time.Unix(7, 0).UTC(), 5.0)

	c := NewCache(512)

	// Write a new value to the cache.
	if err := c.Write("foo", Values{v4, v5}); err != nil {
		t.Fatalf("failed to write value, key foo to cache: %s", err.Error())
	}

	expValues := Values{v5, v4}
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("pre-snapshot write values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}

	snapshots := c.PrepareSnapshots([]string{segment2})

	if err := c.Write("foo", Values{v6}); err != nil {
		t.Fatalf("failed to write post-prepare-snapshots key foo to cache: %s", err.Error())
	}

	expValues = Values{v5, v4, v6}
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("post-prepare-snapshots values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}

	c.RollbackSnapshots(snapshots)

	expValues = Values{v5, v4, v6}
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("post-rollback values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}

	snapshots = c.PrepareSnapshots([]string{segment2, segment3, segment4})
	expLength := 2
	if length := len(snapshots); length != expLength {
		t.Fatalf("post-prepare-snap-shots: length of snaphots incorrect, exp: %v, got %v", expLength, length)
	}

	expFiles := [][]string{[]string{segment2}, []string{segment3, segment4}}
	if files := [][]string{snapshots[0].Files(), snapshots[1].Files()}; !reflect.DeepEqual(expFiles, files) {
		t.Fatalf("files is incorrect, exp: %v, got %v", expFiles, files)
	}

	expFooValues := []Values{Values{v5, v4}, Values{v6}}
	if fooValues := []Values{snapshots[0].Values("foo"), snapshots[1].Values("foo")}; !reflect.DeepEqual(expFooValues, fooValues) {
		t.Fatalf("snapshot values are incorrect, exp: %v, got %v", expFooValues, fooValues)
	}

	snapshots[0] = nil
	c.RollbackSnapshots(snapshots)

	expValues = Values{v6}
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("post-rollback values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}

	snapshots = c.PrepareSnapshots([]string{segment3, segment4, segment5})
	expLength = 2
	if length := len(snapshots); length != expLength {
		t.Fatalf("post-prepare-snap-shots: length of snaphots incorrect, exp: %v, got %v", expLength, length)
	}

	expFiles = [][]string{[]string{segment3, segment4}, []string{segment5}}
	if files := [][]string{snapshots[0].Files(), snapshots[1].Files()}; !reflect.DeepEqual(expFiles, files) {
		t.Fatalf("snapshot files are incorrect, exp: %v, got %v", expFiles, files)
	}

	expFooValues = []Values{Values{v6}, nil}
	if fooValues := []Values{snapshots[0].Values("foo"), snapshots[1].Values("foo")}; !reflect.DeepEqual(expFooValues, fooValues) {
		t.Fatalf("snapshot values are incorrect, exp: %v, got %v", expFooValues, fooValues)
	}

	c.CommitSnapshots()

	expValues = nil
	if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
		t.Fatalf("post-commit values for foo incorrect, exp: %v, got %v", expValues, deduped)
	}
}

// Test that the commit lock prevents two threads preparing snapshots from the same
// cache.
func TestCache_CacheTestCommitLock(t *testing.T) {

	v0 := NewValue(time.Unix(2, 0).UTC(), 0.0)
	in := make(chan int, 0)

	go func() {
		c := NewCache(512)
		out := make(chan int, 1)
		lockTaken := make(chan struct{}, 1)
		valueWritten := make(chan error, 0)

		out <- 0

		_ = c.PrepareSnapshots([]string{"sync"})

		lockTaken <- struct{}{}

		go func() {

			_ = <-lockTaken

			valueWritten <- c.Write("foo", Values{v0})

			c.PrepareSnapshots([]string{"async"})
			observedState := <-out
			c.CommitSnapshots()
			in <- observedState
		}()

		if err := <-valueWritten; err != nil {
			t.Fatalf("concurrent write value failed: %s", err)
		}

		expValues := Values{v0}
		if deduped := c.Values("foo"); !reflect.DeepEqual(expValues, deduped) {
			t.Fatalf("unexpected values in cache after goroutine has written exp: %v, got %v", expValues, deduped)
		}

		select {
		case _ = <-out:
		default:
			t.Fatalf("locking failed in state 0")
		}
		out <- 1

		// give the async goroutine a small chance to see state 1
		time.Sleep(time.Millisecond * 100)

		select {
		case _ = <-out:
		default:
			t.Fatalf("locking failed in state 1")
		}
		out <- 2

		c.CommitSnapshots()
	}()

	expectedState := 2
	select {
	case observedState := <-in:
		if observedState != expectedState {
			t.Fatalf("locking failed: expected observed state: %d, got: %d", expectedState, observedState)
		}
	case _ = <-time.NewTimer(time.Second).C:
		// this should never happen except on really slow machines. consider bumping the value higher.
		t.Fatalf("unexpected deadlock encountered during locking test")
	}

}

func TestCache_CacheEmptySnapshot(t *testing.T) {
	c := NewCache(512)

	// Grab snapshot, and ensure it's as expected.
	snapshots := c.PrepareSnapshots([]string{"foo.wal"})
	if deduped := snapshots[0].Values("foo"); !reflect.DeepEqual(Values(nil), deduped) {
		t.Fatalf("snapshotted values for foo incorrect, exp: %v, got %v", nil, deduped)
	}

	// Ensure cache is still as expected.
	if deduped := c.Values("foo"); !reflect.DeepEqual(Values(nil), deduped) {
		t.Fatalf("post-snapshotted values for foo incorrect, exp: %v, got %v", Values(nil), deduped)
	}

	// Clear snapshot.
	c.CommitSnapshots()
	if deduped := c.Values("foo"); !reflect.DeepEqual(Values(nil), deduped) {
		t.Fatalf("post-snapshot-clear values for foo incorrect, exp: %v, got %v", Values(nil), deduped)
	}
}

func TestCache_CacheWriteMemoryExceeded(t *testing.T) {
	v0 := NewValue(time.Unix(1, 0).UTC(), 1.0)
	v1 := NewValue(time.Unix(2, 0).UTC(), 2.0)

	c := NewCache(uint64(v1.Size()))

	if err := c.Write("foo", Values{v0}); err != nil {
		t.Fatalf("failed to write key foo to cache: %s", err.Error())
	}
	if exp, keys := []string{"foo"}, c.Keys(); !reflect.DeepEqual(keys, exp) {
		t.Fatalf("cache keys incorrect after writes, exp %v, got %v", exp, keys)
	}
	if err := c.Write("bar", Values{v1}); err != ErrCacheMemoryExceeded {
		t.Fatalf("wrong error writing key bar to cache")
	}

	// Grab snapshot, write should still fail since we're still using the memory.
	_ = c.PrepareSnapshots([]string{"foobar.wal"})
	if err := c.Write("bar", Values{v1}); err != ErrCacheMemoryExceeded {
		t.Fatalf("wrong error writing key bar to cache")
	}

	// Clear the snapshot and the write should now succeed.
	c.CommitSnapshots()
	if err := c.Write("bar", Values{v1}); err != nil {
		t.Fatalf("failed to write key foo to cache: %s", err.Error())
	}
	expAscValues := Values{v1}
	if deduped := c.Values("bar"); !reflect.DeepEqual(expAscValues, deduped) {
		t.Fatalf("deduped ascending values for bar incorrect, exp: %v, got %v", expAscValues, deduped)
	}
}

// Ensure the CacheLoader can correctly load from a single segment, even if it's corrupted.
func TestCacheLoader_LoadSingle(t *testing.T) {
	// Create a WAL segment.
	dir := mustTempDir()
	defer os.RemoveAll(dir)
	f := mustTempFile(dir)
	w := NewWALSegmentWriter(f)

	p1 := NewValue(time.Unix(1, 0), 1.1)
	p2 := NewValue(time.Unix(1, 0), int64(1))
	p3 := NewValue(time.Unix(1, 0), true)

	values := map[string][]Value{
		"foo": []Value{p1},
		"bar": []Value{p2},
		"baz": []Value{p3},
	}

	entry := &WriteWALEntry{
		Values: values,
	}

	if err := w.Write(mustMarshalEntry(entry)); err != nil {
		t.Fatal("write points", err)
	}

	// Load the cache using the segment.
	cache := NewCache(1024)
	loader := NewCacheLoader([]string{f.Name()})
	if err := loader.Load(cache); err != nil {
		t.Fatalf("failed to load cache: %s", err.Error())
	}

	// Check the cache.
	if values := cache.Values("foo"); !reflect.DeepEqual(values, Values{p1}) {
		t.Fatalf("cache key foo not as expected, got %v, exp %v", values, Values{p1})
	}
	if values := cache.Values("bar"); !reflect.DeepEqual(values, Values{p2}) {
		t.Fatalf("cache key foo not as expected, got %v, exp %v", values, Values{p2})
	}
	if values := cache.Values("baz"); !reflect.DeepEqual(values, Values{p3}) {
		t.Fatalf("cache key foo not as expected, got %v, exp %v", values, Values{p3})
	}

	// Corrupt the WAL segment.
	if _, err := f.Write([]byte{1, 4, 0, 0, 0}); err != nil {
		t.Fatalf("corrupt WAL segment: %s", err.Error())
	}

	// Reload the cache using the segment.
	cache = NewCache(1024)
	loader = NewCacheLoader([]string{f.Name()})
	if err := loader.Load(cache); err != nil {
		t.Fatalf("failed to load cache: %s", err.Error())
	}

	// Check the cache.
	if values := cache.Values("foo"); !reflect.DeepEqual(values, Values{p1}) {
		t.Fatalf("cache key foo not as expected, got %v, exp %v", values, Values{p1})
	}
	if values := cache.Values("bar"); !reflect.DeepEqual(values, Values{p2}) {
		t.Fatalf("cache key bar not as expected, got %v, exp %v", values, Values{p2})
	}
	if values := cache.Values("baz"); !reflect.DeepEqual(values, Values{p3}) {
		t.Fatalf("cache key baz not as expected, got %v, exp %v", values, Values{p3})
	}
}

// Ensure the CacheLoader can correctly load from two segments, even if one is corrupted.
func TestCacheLoader_LoadDouble(t *testing.T) {
	// Create a WAL segment.
	dir := mustTempDir()
	defer os.RemoveAll(dir)
	f1, f2 := mustTempFile(dir), mustTempFile(dir)
	w1, w2 := NewWALSegmentWriter(f1), NewWALSegmentWriter(f2)

	p1 := NewValue(time.Unix(1, 0), 1.1)
	p2 := NewValue(time.Unix(1, 0), int64(1))
	p3 := NewValue(time.Unix(1, 0), true)
	p4 := NewValue(time.Unix(1, 0), "string")

	// Write first and second segment.

	segmentWrite := func(w *WALSegmentWriter, values map[string][]Value) {
		entry := &WriteWALEntry{
			Values: values,
		}
		if err := w1.Write(mustMarshalEntry(entry)); err != nil {
			t.Fatal("write points", err)
		}
	}

	values := map[string][]Value{
		"foo": []Value{p1},
		"bar": []Value{p2},
	}
	segmentWrite(w1, values)
	values = map[string][]Value{
		"baz": []Value{p3},
		"qux": []Value{p4},
	}
	segmentWrite(w2, values)

	// Corrupt the first WAL segment.
	if _, err := f1.Write([]byte{1, 4, 0, 0, 0}); err != nil {
		t.Fatalf("corrupt WAL segment: %s", err.Error())
	}

	// Load the cache using the segments.
	cache := NewCache(1024)
	loader := NewCacheLoader([]string{f1.Name(), f2.Name()})
	if err := loader.Load(cache); err != nil {
		t.Fatalf("failed to load cache: %s", err.Error())
	}

	// Check the cache.
	if values := cache.Values("foo"); !reflect.DeepEqual(values, Values{p1}) {
		t.Fatalf("cache key foo not as expected, got %v, exp %v", values, Values{p1})
	}
	if values := cache.Values("bar"); !reflect.DeepEqual(values, Values{p2}) {
		t.Fatalf("cache key bar not as expected, got %v, exp %v", values, Values{p2})
	}
	if values := cache.Values("baz"); !reflect.DeepEqual(values, Values{p3}) {
		t.Fatalf("cache key baz not as expected, got %v, exp %v", values, Values{p3})
	}
	if values := cache.Values("qux"); !reflect.DeepEqual(values, Values{p4}) {
		t.Fatalf("cache key qux not as expected, got %v, exp %v", values, Values{p4})
	}
}

func mustTempDir() string {
	dir, err := ioutil.TempDir("", "tsm1-test")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp dir: %v", err))
	}
	return dir
}

func mustTempFile(dir string) *os.File {
	f, err := ioutil.TempFile(dir, "tsm1test")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp file: %v", err))
	}
	return f
}

func mustMarshalEntry(entry WALEntry) (WalEntryType, []byte) {
	bytes := make([]byte, 1024<<2)

	b, err := entry.Encode(bytes)
	if err != nil {
		panic(fmt.Sprintf("error encoding: %v", err))
	}

	return entry.Type(), snappy.Encode(b, b)
}
