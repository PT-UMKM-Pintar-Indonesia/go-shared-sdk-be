package sdk_helper

import (
	"reflect"
	"sort"
	"testing"
)

func TestSliceSearchAndDedup(t *testing.T) {
	if !Contains([]int{1, 2, 3}, 2) || Contains([]int{1}, 9) {
		t.Fatal("Contains")
	}

	if got := Unique([]int{1, 1, 2, 3, 3}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Unique got %v", got)
	}

	if got := Uniq([]string{"a", "a", "b"}); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("Uniq got %v", got)
	}

	if got := UniqBy([]int{1, 2, 3, 4}, func(n int) int { return n % 2 }); len(got) != 2 {
		t.Fatalf("UniqBy got %v", got)
	}

	if got := UniqWith([]int{1, 2, 3, 4}, func(a, b int) bool { return a%3 == b%3 }); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("UniqWith got %v", got)
	}

	if got := UniqWith([]int{}, func(a, b int) bool { return a == b }); len(got) != 0 {
		t.Fatal("UniqWith empty")
	}
}

func TestSliceHeadTail(t *testing.T) {
	if v, ok := Head([]int{5, 6}); !ok || v != 5 {
		t.Fatal("Head")
	}

	if _, ok := Head([]int{}); ok {
		t.Fatal("Head empty")
	}

	if v, ok := Last([]int{5, 6}); !ok || v != 6 {
		t.Fatal("Last")
	}

	if _, ok := Last([]string{}); ok {
		t.Fatal("Last empty")
	}

	if v, ok := At([]int{1, 2, 3}, -1); !ok || v != 3 {
		t.Fatalf("At negative %v %v", v, ok)
	}

	if _, ok := At([]int{1}, 5); ok {
		t.Fatal("At out of range")
	}
}

func TestSliceSetOps(t *testing.T) {
	if got := Union([]int{1, 2}, []int{2, 3}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Union got %v", got)
	}

	inter := Intersection([]int{1, 2, 3}, []int{2, 3, 4})
	sort.Ints(inter)

	if len(inter) != 2 || inter[0] != 2 || inter[1] != 3 {
		t.Fatalf("Intersection got %v", inter)
	}

	if got := Intersection[int](); len(got) != 0 {
		t.Fatal("Intersection no args")
	}

	if got := Intersection([]int{1, 1, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Intersection single slice should dedup, got %v", got)
	}

	xor := Xor([]int{1, 2}, []int{2, 3})
	sort.Ints(xor)

	if len(xor) != 2 || xor[0] != 1 || xor[1] != 3 {
		t.Fatalf("Xor got %v", xor)
	}

	if got := Difference([]int{1, 2, 3}, []int{2}, []int{3}); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("Difference got %v", got)
	}

	if got := DifferenceBy([]int{1, 2, 3}, []int{4}, func(n int) int { return n % 3 }); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Fatalf("DifferenceBy got %v", got)
	}

	if got := DifferenceWith([]int{1, 2, 3}, []int{9}, func(a, b int) bool { return a == b }); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("DifferenceWith got %v", got)
	}

	if got := Pull([]int{1, 2, 3, 2}, 2); !reflect.DeepEqual(got, []int{1, 3}) {
		t.Fatalf("Pull got %v", got)
	}

	if got := PullAll([]int{1, 2}, []int{1}); !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("PullAll got %v", got)
	}

	if got := PullAllBy([]int{1, 2, 3}, []int{4}, func(n int) int { return n % 3 }); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Fatalf("PullAllBy got %v", got)
	}

	if got := PullAllWith([]int{1, 2}, []int{5}, func(a, b int) bool { return a == b }); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("PullAllWith got %v", got)
	}

	if got := Take([]int{1, 2, 3}, 2); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Take got %v", got)
	}

	if got := Take([]int{1}, 0); len(got) != 0 {
		t.Fatal("Take zero")
	}

	if got := Truncate([]int{1, 2, 3}, 10); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatal("Truncate beyond length")
	}
}

func TestSliceTransforms(t *testing.T) {
	if got := Compact([]int{0, 1, 0, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Compact got %v", got)
	}

	chunks := Chunk([]int{1, 2, 3, 4, 5}, 2)
	if len(chunks) != 3 || len(chunks[2]) != 1 {
		t.Fatalf("Chunk got %v", chunks)
	}

	if got := Chunk([]int{1}, 0); len(got) != 0 {
		t.Fatal("Chunk zero size")
	}

	grouped := GroupBy([]int{1, 2, 3, 4}, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})

	if len(grouped["even"]) != 2 || len(grouped["odd"]) != 2 {
		t.Fatalf("GroupBy got %v", grouped)
	}

	t1, f1 := Partition([]int{1, 2, 3}, func(n int) bool { return n > 1 })
	if len(t1) != 2 || len(f1) != 1 {
		t.Fatalf("Partition got %v %v", t1, f1)
	}

	if got := Map([]int{1, 2}, func(n int) int { return n * 2 }); !reflect.DeepEqual(got, []int{2, 4}) {
		t.Fatalf("Map got %v", got)
	}

	if got := Flat([][]int{{1}, {2, 3}}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Flat got %v", got)
	}

	if got := FlatMap([]int{1, 2}, func(n int) []int { return []int{n, n} }); len(got) != 4 {
		t.Fatalf("FlatMap got %v", got)
	}

	if got := Concat([]int{1}, []int{2}, []int{3}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Concat got %v", got)
	}

	if got := Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 }); !reflect.DeepEqual(got, []int{2, 4}) {
		t.Fatalf("Filter got %v", got)
	}

	if got := Of(1, 2); len(got) != 2 || got[0] != 1 {
		t.Fatal("Of")
	}

	if got := From("a"); got[0] != "a" {
		t.Fatal("From")
	}

	if got := Unshift([]int{3}, 1, 2); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Unshift got %v", got)
	}

	if got := Push([]int{1}, 2); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatal("Push")
	}

	if got := Join([]string{"a", "b"}, "-"); got != "a-b" {
		t.Fatal("Join")
	}

	if got := ToString([]int{1, 2}); got != "[1, 2]" {
		t.Fatalf("ToString got %q", got)
	}

	entries := Entries([]string{"x", "y"})
	if len(entries) != 2 || entries[1].Index != 1 || entries[1].Value != "y" {
		t.Fatalf("Entries got %v", entries)
	}
}

func TestSlicePredicates(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	if !Every([]int{2, 4}, isEven) || Every([]int{2, 3}, isEven) {
		t.Fatal("Every")
	}

	if !Some([]int{1, 2}, isEven) || Some([]int{1, 3}, isEven) {
		t.Fatal("Some")
	}

	if v, ok := Find([]int{1, 2, 3}, isEven); !ok || v != 2 {
		t.Fatal("Find")
	}

	if _, ok := Find([]int{1}, isEven); ok {
		t.Fatal("Find miss")
	}

	if v, ok := FindLast([]int{1, 2, 3, 4}, isEven); !ok || v != 4 {
		t.Fatal("FindLast")
	}

	if i := FindIndex([]int{1, 2}, isEven); i != 1 {
		t.Fatal("FindIndex")
	}

	if i := FindLastIndex([]int{2, 4}, isEven); i != 1 {
		t.Fatal("FindLastIndex")
	}

	if i := FindIndex([]int{1}, isEven); i != -1 {
		t.Fatal("FindIndex miss must be -1")
	}

	if !Includes([]int{1, 2}, 2) || Includes([]int{1}, 5) {
		t.Fatal("Includes")
	}

	if IndexOf([]int{5, 6}, 6) != 1 || IndexOf([]int{5}, 9) != -1 {
		t.Fatal("IndexOf")
	}

	if LastIndexOf([]int{1, 2, 1}, 1, 2) != 2 {
		t.Fatal("LastIndexOf")
	}

	if LastIndexOf([]int{1, 2, 1}, 1, -1) != -1 {
		t.Fatal("LastIndexOf negative")
	}
}

func TestSliceMutationStyle(t *testing.T) {
	src := []int{1, 2, 3, 4, 5}

	if got := Select(src, 1, 3); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Fatalf("Select got %v", got)
	}

	if got := Select(src, 3, 1); len(got) != 0 {
		t.Fatal("Select inverted")
	}

	if got := Select(src, 0, 100); len(got) != 5 {
		t.Fatal("Select clamps end")
	}

	filled := Fill([]int{1, 2, 3}, 0, 1, 2)
	if !reflect.DeepEqual(filled, []int{1, 0, 3}) {
		t.Fatalf("Fill got %v", filled)
	}

	if got := Fill([]int{1, 2}, 0, 5, 9); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatal("Fill out of range no-op")
	}

	rev := Reverse([]int{1, 2, 3})
	if !reflect.DeepEqual(rev, []int{3, 2, 1}) {
		t.Fatal("Reverse")
	}

	orig := []int{3, 1, 2}
	safe := ToSorted(orig)

	if !reflect.DeepEqual(safe, []int{1, 2, 3}) {
		t.Fatalf("ToSorted got %v", safe)
	}

	if !reflect.DeepEqual(orig, []int{3, 1, 2}) {
		t.Fatal("ToSorted must not mutate input")
	}

	if got := ToReversed([]int{1, 2}); !reflect.DeepEqual(got, []int{2, 1}) {
		t.Fatal("ToReversed")
	}

	if got := ToSortedFunc([]int{3, 1, 2}, func(a, b int) bool { return a > b }); !reflect.DeepEqual(got, []int{3, 2, 1}) {
		t.Fatalf("ToSortedFunc got %v", got)
	}

	w := With([]int{1, 2, 3}, 1, 9)
	if !reflect.DeepEqual(w, []int{1, 9, 3}) {
		t.Fatalf("With got %v", w)
	}

	if got := With([]int{1}, 5, 2); !reflect.DeepEqual(got, []int{1}) {
		t.Fatal("With out of range no-op")
	}

	removed, rest := Splice([]int{1, 2, 3, 4}, 1, 2, 9)
	if !reflect.DeepEqual(removed, []int{2, 3}) || !reflect.DeepEqual(rest, []int{1, 9, 4}) {
		t.Fatalf("Splice got %v %v", removed, rest)
	}

	if got := ToSpliced([]int{1, 2, 3}, -2, 1); !reflect.DeepEqual(got, []int{1, 3}) {
		t.Fatalf("ToSpliced negative start got %v", got)
	}

	cw := CopyWithin([]int{1, 2, 3, 4, 5}, 0, 3, 5)
	if !reflect.DeepEqual(cw, []int{4, 5, 3, 4, 5}) {
		t.Fatalf("CopyWithin got %v", cw)
	}

	v, tail, ok := Pop([]int{1, 2, 3})
	if !ok || v != 3 || len(tail) != 2 {
		t.Fatal("Pop")
	}

	h, head, ok := Shift([]int{1, 2, 3})
	if !ok || h != 1 || len(head) != 2 {
		t.Fatal("Shift")
	}

	if _, _, ok := Pop[int](nil); ok {
		t.Fatal("Pop empty")
	}

	if _, _, ok := Shift[int](nil); ok {
		t.Fatal("Shift empty")
	}
}

func TestSliceReduceAndSort(t *testing.T) {
	sum := Reduce([]int{1, 2, 3}, 0, func(acc, n int) int { return acc + n })
	if sum != 6 {
		t.Fatalf("Reduce got %d", sum)
	}

	joint := ReduceRight([]string{"a", "b"}, "", func(acc, s string) string { return acc + s })
	if joint != "ba" {
		t.Fatalf("ReduceRight got %q", joint)
	}

	sorted := Sort([]int{3, 1, 2})
	if !reflect.DeepEqual(sorted, []int{1, 2, 3}) {
		t.Fatal("Sort")
	}

	strs := SortFunc([]string{"bb", "a"}, func(x, y string) bool { return len(x) < len(y) })
	if strs[0] != "a" {
		t.Fatal("SortFunc")
	}

	if !IsSlice([]int{}) || IsSlice(1) {
		t.Fatal("IsSlice")
	}

	var count int
	ForEach([]int{1, 2, 3}, func(n int) { count += n })

	if count != 6 {
		t.Fatalf("ForEach sum %d", count)
	}
}
