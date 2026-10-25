package sdk_helper

import (
	"strings"
	"sync"
	"testing"
)

func pair[K comparable, V any](k K, v V) struct {
	Key   K
	Value V
} {
	return struct {
		Key   K
		Value V
	}{Key: k, Value: v}
}

func TestMapSafeBasics(t *testing.T) {
	m := NewMapSafe(pair("a", 1), pair("b", 2))

	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Fatal("Get")
	}

	m.Set("c", 3)
	if !m.Has("c") || m.Size() != 3 {
		t.Fatal("Set/Has/Size")
	}

	if !m.Delete("c") || m.Delete("c") {
		t.Fatal("Delete existence")
	}

	if len(m.Keys()) != 2 || len(m.Values()) != 2 || len(m.Entries()) != 2 {
		t.Fatal("Keys/Values/Entries")
	}

	if !strings.HasPrefix(m.String(), "Map{") {
		t.Fatalf("String got %q", m.String())
	}

	if m.ValueOf() != m {
		t.Fatal("ValueOf must return receiver")
	}

	seen := 0
	m.ForEach(func(string, int) { seen++ })

	if seen != 2 {
		t.Fatal("ForEach")
	}

	m.Clear()
	if m.Size() != 0 {
		t.Fatal("Clear")
	}
}

func TestMapSafeConcurrent(t *testing.T) {
	m := NewMapSafe[int, int]()

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				m.Set(i, j)
				m.Get(i)
				m.Has(i)
			}
		}(i)
	}

	wg.Wait()

	if m.Size() != 8 {
		t.Fatalf("expected 8 keys, got %d", m.Size())
	}
}

func TestMapUnSafeBasics(t *testing.T) {
	m := NewMapUnSafe(pair("x", 10))

	m.Set("y", 20)

	if v, ok := m.Get("y"); !ok || v != 20 {
		t.Fatal("Get")
	}

	if !m.Has("x") || m.Size() != 2 {
		t.Fatal("Has/Size")
	}

	if !m.Delete("x") || m.Delete("x") {
		t.Fatal("Delete")
	}

	if len(m.Keys()) != 1 || len(m.Values()) != 1 || len(m.Entries()) != 1 {
		t.Fatal("Keys/Values/Entries")
	}

	seen := 0
	m.ForEach(func(string, int) { seen++ })

	if seen != 1 {
		t.Fatal("ForEach")
	}

	if !strings.HasPrefix(m.String(), "Map{") {
		t.Fatalf("String %q", m.String())
	}

	if m.ValueOf() != m {
		t.Fatal("ValueOf")
	}

	m.Clear()
	if m.Size() != 0 {
		t.Fatal("Clear")
	}
}
