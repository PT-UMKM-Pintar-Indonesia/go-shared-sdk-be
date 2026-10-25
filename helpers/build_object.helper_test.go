package sdk_helper

import (
	"sort"
	"strings"
	"sync"
	"testing"
)

func TestKeysValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	ks := Keys(m)
	sort.Strings(ks)

	if len(ks) != 2 || ks[0] != "a" || ks[1] != "b" {
		t.Fatalf("Keys got %v", ks)
	}

	vs := Values(m)
	sort.Ints(vs)

	if len(vs) != 2 || vs[0] != 1 || vs[1] != 2 {
		t.Fatalf("Values got %v", vs)
	}
}

func TestMapObjectSafeCRUD(t *testing.T) {
	o := NewMapObjectSafe(map[string]int{"seed": 1})

	if v, ok := o.Get("seed"); !ok || v != 1 {
		t.Fatal("initial seed missing")
	}

	o.Set("a", 10)
	o.Assign(map[string]int{"b": 20, "c": 30})

	if !o.Has("b") || o.Has("zz") {
		t.Fatal("Has")
	}

	o.Delete("c")
	if o.Has("c") {
		t.Fatal("Delete")
	}

	if len(o.Keys()) != 3 {
		t.Fatalf("Keys got %v", o.Keys())
	}

	if sum := o.Values(); len(sum) != 3 {
		t.Fatalf("Values got %v", sum)
	}

	if len(o.Entries()) != 3 {
		t.Fatal("Entries")
	}

	if !strings.HasPrefix(o.String(), "MapObjectSafe{") {
		t.Fatalf("String got %q", o.String())
	}

	// Mutating the source map must not affect the object (copy semantics).
	src := map[string]int{"x": 1}
	obj := NewMapObjectSafe(src)
	src["y"] = 2
	if obj.Has("y") {
		t.Fatal("expected defensive copy on construction")
	}
}

func TestMapObjectSafeConcurrent(t *testing.T) {
	o := NewMapObjectSafe[string, int](nil)

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				o.Set(strings.Repeat("k", i), j)
				_ = o.Keys()
				_, _ = o.Get(strings.Repeat("k", i))
			}
		}(i)
	}

	wg.Wait()

	if len(o.Keys()) != 8 {
		t.Fatalf("expected 8 keys, got %d", len(o.Keys()))
	}
}

func TestMapObjectUnSafeCRUD(t *testing.T) {
	o := NewMapObjectUnSafe(map[string]int{"seed": 1})

	o.Set("a", 2)
	o.Assign(map[string]int{"b": 3})
	o.Delete("seed")

	if o.Has("seed") || !o.Has("a") || !o.Has("b") {
		t.Fatal("unexpected state")
	}

	if v, ok := o.Get("b"); !ok || v != 3 {
		t.Fatal("Get")
	}

	if len(o.Entries()) != 2 || len(o.Values()) != 2 {
		t.Fatal("Entries/Values")
	}

	if !strings.HasPrefix(o.String(), "MapObjectUnSafe{") {
		t.Fatalf("String got %q", o.String())
	}
}
