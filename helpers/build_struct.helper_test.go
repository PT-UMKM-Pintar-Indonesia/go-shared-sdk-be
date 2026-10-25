package sdk_helper

import (
	"testing"
)

type structSample struct {
	Name  string
	Count int
	priv  string
}

func TestNewStructObjectRejectsNonPointer(t *testing.T) {
	if _, err := NewStructObject(structSample{}); err == nil {
		t.Fatal("expected error for non-pointer")
	}

	if _, err := NewStructObject(42); err == nil {
		t.Fatal("expected error for non-struct")
	}
}

func TestStructObjectGetSet(t *testing.T) {
	s := &structSample{Name: "init", priv: "hidden"}

	obj, err := NewStructObject(s)
	if err != nil {
		t.Fatal(err)
	}

	if v, ok := obj.Get("Name"); !ok || v != "init" {
		t.Fatalf("Get Name: %v %v", v, ok)
	}

	if _, ok := obj.Get("Missing"); ok {
		t.Fatal("expected miss for unknown field")
	}

	if err := obj.Set("Count", 5); err != nil {
		t.Fatal(err)
	}

	if s.Count != 5 {
		t.Fatal("Set did not mutate underlying struct")
	}

	if err := obj.Set("Count", "wrong type"); err == nil {
		t.Fatal("expected type mismatch error")
	}

	if err := obj.Set("Missing", 1); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestStructObjectKeysSkipsUnexported(t *testing.T) {
	obj, err := NewStructObject(&structSample{})
	if err != nil {
		t.Fatal(err)
	}

	keys := obj.Keys()

	for _, k := range keys {
		if k == "priv" {
			t.Fatalf("unexported field %q must be excluded", k)
		}
	}

	if len(keys) != 2 {
		t.Fatalf("expected 2 exported keys, got %v", keys)
	}
}

func TestStructObjectAssign(t *testing.T) {
	obj, err := NewStructObject(&structSample{})
	if err != nil {
		t.Fatal(err)
	}

	if err := obj.Assign(structSample{Name: "assigned", Count: 9}); err != nil {
		t.Fatal(err)
	}

	if v, _ := obj.Get("Name"); v != "assigned" {
		t.Fatalf("Assign Name got %v", v)
	}

	if err := obj.Assign(42); err == nil {
		t.Fatal("expected error assigning from non-struct")
	}
}

func TestStructObjectFieldCache(t *testing.T) {
	// Second lookup must hit the sync.Map cache and return the same index.
	obj, _ := NewStructObject(&structSample{})

	i1, ok1 := obj.getFieldIndex("Count")
	i2, ok2 := obj.getFieldIndex("Count")

	if !ok1 || !ok2 || i1 != i2 || i1 != 1 {
		t.Fatalf("cache inconsistency: %d/%v %d/%v", i1, ok1, i2, ok2)
	}
}
