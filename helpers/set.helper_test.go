package sdk_helper

import (
	"sort"
	"sync"
	"testing"
)

func sortedInts(in []int) []int {
	out := append([]int(nil), in...)
	sort.Ints(out)
	return out
}

func TestSetSafeBasics(t *testing.T) {
	s := NewSetSafe(1, 2, 2, 3)

	if s.Size() != 3 {
		t.Fatalf("expected dedup on construction, size %d", s.Size())
	}

	s.Add(4)
	if !s.Has(4) {
		t.Fatal("Add/Has")
	}

	if !s.Delete(4) || s.Delete(4) {
		t.Fatal("Delete must report existence")
	}

	s.ForEach(func(el int) {
		if el < 1 || el > 3 {
			t.Errorf("unexpected element %d", el)
		}
	})

	if len(s.Entries()) != 3 {
		t.Fatal("Entries")
	}

	s.Clear()
	if s.Size() != 0 {
		t.Fatal("Clear")
	}
}

func TestSetSafeAlgebra(t *testing.T) {
	a := NewSetSafe(1, 2, 3, 4)
	b := NewSetSafe(3, 4, 5, 6)

	if got := sortedInts(a.Union(b).Values()); len(got) != 6 {
		t.Fatalf("Union got %v", got)
	}

	inter := sortedInts(a.Intersection(b).Values())
	if len(inter) != 2 || inter[0] != 3 || inter[1] != 4 {
		t.Fatalf("Intersection got %v", inter)
	}

	diff := sortedInts(a.Difference(b).Values())
	if len(diff) != 2 || diff[0] != 1 || diff[1] != 2 {
		t.Fatalf("Difference got %v", diff)
	}

	sym := sortedInts(a.SymmetricDifference(b).Values())
	if len(sym) != 4 {
		t.Fatalf("SymmetricDifference got %v", sym)
	}

	small := NewSetSafe(1, 2)
	if !small.IsSubsetOf(a) || a.IsSubsetOf(small) {
		t.Fatal("IsSubsetOf")
	}

	if !a.IsSupersetOf(small) {
		t.Fatal("IsSupersetOf")
	}

	if !NewSetSafe(7).IsDisjointFrom(a) || a.IsDisjointFrom(b) {
		t.Fatal("IsDisjointFrom")
	}
}

func TestSetSafeConcurrent(t *testing.T) {
	s := NewSetSafe[int]()

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s.Add(i*1000 + j)
				s.Has(i)
			}
		}(i)
	}

	wg.Wait()

	if s.Size() != 400 {
		t.Fatalf("expected 400 elements, got %d", s.Size())
	}
}

func TestSetUnSafeBasics(t *testing.T) {
	s := NewSetUnSafe("a", "b")

	s.Add("c")
	if s.Size() != 3 || !s.Has("c") {
		t.Fatal("Add/Has/Size")
	}

	if !s.Delete("a") || s.Delete("a") {
		t.Fatal("Delete")
	}

	u := s.Union(NewSetUnSafe("z"))
	if u.Size() != 3 {
		t.Fatalf("Union got %d", u.Size())
	}

	i := s.Intersection(NewSetUnSafe("b", "z"))
	if i.Size() != 1 || !i.Has("b") {
		t.Fatal("Intersection")
	}

	d := s.Difference(NewSetUnSafe("b"))
	if d.Size() != 1 || !d.Has("c") {
		t.Fatal("Difference")
	}

	sd := s.SymmetricDifference(NewSetUnSafe("c", "q"))
	if sd.Size() != 2 || !sd.Has("b") || !sd.Has("q") {
		t.Fatalf("SymmetricDifference got %v", sd.Values())
	}

	if !s.IsSubsetOf(NewSetUnSafe("b", "c", "extra")) {
		t.Fatal("IsSubsetOf")
	}

	if s.IsDisjointFrom(NewSetUnSafe("b")) {
		t.Fatal("IsDisjointFrom")
	}

	var count int
	s.ForEach(func(string) { count++ })

	if count != 2 || len(s.Entries()) != 2 || len(s.Values()) != 2 {
		t.Fatal("ForEach/Entries/Values")
	}

	s.Clear()
	if s.Size() != 0 {
		t.Fatal("Clear")
	}
}
