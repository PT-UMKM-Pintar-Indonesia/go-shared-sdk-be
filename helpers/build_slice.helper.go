package sdk_helper

import (
	"cmp"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type (
	Predicate[T any] func(T) bool

	Iteratee[T any, K comparable] func(T) K

	Comparator[T any] func(T, T) bool
)

func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func Unique[T comparable](slice []T) []T {
	keys := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))

	for _, entry := range slice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = struct{}{}
			result = append(result, entry)
		}
	}

	return result
}

func Head[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	return slice[0], true
}

func Last[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	return slice[len(slice)-1], true
}

func Uniq[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))

	for _, entry := range slice {
		if _, ok := seen[entry]; !ok {
			seen[entry] = struct{}{}
			result = append(result, entry)
		}
	}

	return result
}

func UniqBy[T any, K comparable](slice []T, iteratee Iteratee[T, K]) []T {
	seen := make(map[K]struct{}, len(slice))
	result := make([]T, 0, len(slice))

	for _, entry := range slice {
		key := iteratee(entry)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, entry)
		}
	}

	return result
}

func UniqWith[T any](slice []T, comparator Comparator[T]) []T {
	if len(slice) == 0 {
		return []T{}
	}

	result := make([]T, 0, len(slice))
	result = append(result, slice[0])

	for i := 1; i < len(slice); i++ {
		isUnique := true
		for _, resElem := range result {
			if comparator(slice[i], resElem) {
				isUnique = false
				break
			}
		}

		if isUnique {
			result = append(result, slice[i])
		}
	}

	return result
}

func Union[T comparable](slices ...[]T) []T {
	return Uniq(Concat(slices...))
}

func UnionBy[T any, K comparable](slices [][]T, iteratee Iteratee[T, K]) []T {
	combined := Concat(slices...)
	return UniqBy(combined, iteratee)
}

func UnionWith[T any](slices [][]T, comparator Comparator[T]) []T {
	combined := Concat(slices...)
	return UniqWith(combined, comparator)
}

func Intersection[T comparable](slices ...[]T) []T {
	if len(slices) == 0 {
		return []T{}
	}

	if len(slices) == 1 {
		return Uniq(slices[0])
	}

	counts := make(map[T]int)
	for _, slice := range slices {
		seenInSlice := make(map[T]struct{})
		for _, item := range slice {
			if _, ok := seenInSlice[item]; !ok {
				counts[item]++
				seenInSlice[item] = struct{}{}
			}
		}
	}

	result := make([]T, 0)
	for item, count := range counts {
		if count == len(slices) {
			result = append(result, item)
		}
	}
	return result
}

func IntersectionBy[T any, K comparable](slices [][]T, iteratee Iteratee[T, K]) []T {
	if len(slices) == 0 {
		return []T{}
	}
	if len(slices) == 1 {
		return UniqBy(slices[0], iteratee)
	}

	type itemInfo struct {
		original T
		count    int
	}
	counts := make(map[K]itemInfo)

	for _, slice := range slices {
		seenInSlice := make(map[K]struct{})
		for _, item := range slice {
			key := iteratee(item)
			if _, ok := seenInSlice[key]; !ok {
				info := counts[key]
				if info.count == 0 {
					info.original = item
				}
				info.count++
				counts[key] = info
				seenInSlice[key] = struct{}{}
			}
		}
	}

	result := make([]T, 0)
	for _, info := range counts {
		if info.count == len(slices) {
			result = append(result, info.original)
		}
	}

	return result
}

func IntersectionWith[T any](slices [][]T, comparator Comparator[T]) []T {
	if len(slices) == 0 {
		return []T{}
	}
	if len(slices) == 1 {
		return UniqWith(slices[0], comparator)
	}

	base := UniqWith(slices[0], comparator)
	result := make([]T, 0, len(base))

	for _, baseItem := range base {
		isInAll := true
		for i := 1; i < len(slices); i++ {
			foundInSlice := false
			for _, otherItem := range slices[i] {
				if comparator(baseItem, otherItem) {
					foundInSlice = true
					break
				}
			}
			if !foundInSlice {
				isInAll = false
				break
			}
		}

		if isInAll {
			result = append(result, baseItem)
		}
	}

	return result
}

func Pull[T comparable](slice []T, values ...T) []T {
	if len(slice) == 0 || len(values) == 0 {
		return slice
	}

	removeMap := make(map[T]struct{}, len(values))
	for _, v := range values {
		removeMap[v] = struct{}{}
	}

	newSlice := make([]T, 0, len(slice))
	for _, item := range slice {
		if _, found := removeMap[item]; !found {
			newSlice = append(newSlice, item)
		}
	}

	return newSlice
}

func PullAll[T comparable](slice []T, values []T) []T {
	return Pull(slice, values...)
}

func PullAllBy[T any, K comparable](slice []T, values []T, iteratee Iteratee[T, K]) []T {
	if len(slice) == 0 || len(values) == 0 {
		return slice
	}

	removeMap := make(map[K]struct{}, len(values))
	for _, v := range values {
		removeMap[iteratee(v)] = struct{}{}
	}

	newSlice := make([]T, 0, len(slice))
	for _, item := range slice {
		if _, found := removeMap[iteratee(item)]; !found {
			newSlice = append(newSlice, item)
		}
	}

	return newSlice
}

func PullAllWith[T any](slice []T, values []T, comparator Comparator[T]) []T {
	if len(slice) == 0 || len(values) == 0 {
		return slice
	}

	newSlice := make([]T, 0, len(slice))
	for _, item := range slice {
		shouldKeep := true
		for _, removeVal := range values {
			if comparator(item, removeVal) {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			newSlice = append(newSlice, item)
		}
	}

	return newSlice
}

func Take[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}

	if n >= len(slice) {
		return slice
	}

	return slice[:n]
}

func Xor[T comparable](slices ...[]T) []T {
	counts := make(map[T]int)
	for _, slice := range slices {
		seenInSlice := make(map[T]struct{})
		for _, item := range slice {
			if _, ok := seenInSlice[item]; !ok {
				counts[item]++
				seenInSlice[item] = struct{}{}
			}
		}
	}

	result := make([]T, 0)
	for item, count := range counts {
		if count%2 != 0 {
			result = append(result, item)
		}
	}

	return result
}

func XorBy[T any, K comparable](slices [][]T, iteratee Iteratee[T, K]) []T {
	type itemInfo struct {
		original T
		count    int
	}

	counts := make(map[K]itemInfo)

	for _, slice := range slices {
		seenInSlice := make(map[K]struct{})
		for _, item := range slice {
			key := iteratee(item)
			if _, ok := seenInSlice[key]; !ok {
				info := counts[key]
				if info.count == 0 {
					info.original = item
				}
				info.count++
				counts[key] = info
				seenInSlice[key] = struct{}{}
			}
		}
	}

	result := make([]T, 0)
	for _, info := range counts {
		if info.count%2 != 0 {
			result = append(result, info.original)
		}
	}

	return result
}

func XorWith[T any](slices [][]T, comparator Comparator[T]) []T {
	if len(slices) == 0 {
		return []T{}
	}

	type entry struct {
		item  T
		count int
	}

	uniqueItems := make([]entry, 0)
	for _, slice := range slices {
		for _, item := range slice {
			found := false
			for i := range uniqueItems {
				if comparator(item, uniqueItems[i].item) {
					uniqueItems[i].count++
					found = true
					break
				}
			}
			if !found {
				uniqueItems = append(uniqueItems, entry{item: item, count: 1})
			}
		}
	}

	result := make([]T, 0, len(uniqueItems))
	for _, e := range uniqueItems {
		if e.count%2 != 0 {
			result = append(result, e.item)
		}
	}

	return result
}

func Compact[T comparable](slice []T) []T {
	var zero T
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		if item != zero {
			result = append(result, item)
		}
	}

	return result
}

func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	if len(slice) == 0 {
		return [][]T{}
	}

	numChunks := (len(slice) + size - 1) / size
	result := make([][]T, 0, numChunks)

	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}

	return result
}

func Difference[T comparable](array []T, values ...[]T) []T {
	removeMap := make(map[T]struct{})

	for _, valSlice := range values {
		for _, val := range valSlice {
			removeMap[val] = struct{}{}
		}
	}

	result := make([]T, 0, len(array))
	for _, item := range array {
		if _, found := removeMap[item]; !found {
			result = append(result, item)
		}
	}

	return result
}

func DifferenceBy[T any, K comparable](array []T, values []T, iteratee Iteratee[T, K]) []T {
	removeMap := make(map[K]struct{})

	for _, val := range values {
		removeMap[iteratee(val)] = struct{}{}
	}

	result := make([]T, 0, len(array))
	for _, item := range array {
		if _, found := removeMap[iteratee(item)]; !found {
			result = append(result, item)
		}
	}

	return result
}

func DifferenceWith[T any](array []T, values []T, comparator Comparator[T]) []T {
	result := make([]T, 0, len(array))

	for _, item := range array {
		shouldKeep := true

		for _, removeVal := range values {
			if comparator(item, removeVal) {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			result = append(result, item)
		}
	}

	return result
}

func GroupBy[T any, K comparable](collection []T, iteratee Iteratee[T, K]) map[K][]T {
	result := make(map[K][]T)

	for _, item := range collection {
		key := iteratee(item)
		result[key] = append(result[key], item)
	}

	return result
}

func Partition[T any](collection []T, predicate Predicate[T]) ([]T, []T) {
	trueSlice := make([]T, 0, len(collection)/2)
	falseSlice := make([]T, 0, len(collection)/2)

	for _, item := range collection {
		if predicate(item) {
			trueSlice = append(trueSlice, item)
		} else {
			falseSlice = append(falseSlice, item)
		}
	}

	return trueSlice, falseSlice
}

func Truncate[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return slice
	}

	return slice[:n]
}

func Of[T any](elements ...T) []T {
	return elements
}

func From[T any](elements ...T) []T {
	return elements
}

func At[T any](slice []T, index int) (T, bool) {
	length := len(slice)
	if index < 0 {
		index = length + index
	}

	if index < 0 || index >= length {
		var zero T
		return zero, false
	}

	return slice[index], true
}

func Concat[T any](slices ...[]T) []T {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, 0, totalLen)

	for _, s := range slices {
		result = append(result, s...)
	}

	return result
}

func CopyWithin[T any](slice []T, target, start, end int) []T {
	length := len(slice)
	if target < 0 {
		target = length + target
	}

	if start < 0 {
		start = length + start
	}

	if end < 0 {
		end = length + end
	}

	if target < 0 {
		target = 0
	}

	if start < 0 {
		start = 0
	}

	if end > length {
		end = length
	}

	if target >= length || start >= length || start >= end {
		return slice
	}

	copyLen := end - start
	if target+copyLen > length {
		copyLen = length - target
	}

	copy(slice[target:target+copyLen], slice[start:start+copyLen])

	return slice
}

func Entries[T any](slice []T) []struct {
	Index int
	Value T
} {
	entries := make([]struct {
		Index int
		Value T
	}, len(slice))

	for i, v := range slice {
		entries[i] = struct {
			Index int
			Value T
		}{Index: i, Value: v}
	}

	return entries
}

func Every[T any](slice []T, predicate Predicate[T]) bool {
	for _, v := range slice {
		if !predicate(v) {
			return false
		}
	}

	return true
}

func Fill[T any](slice []T, value T, start, end int) []T {
	length := len(slice)
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start >= end {
		return slice
	}

	for i := start; i < end; i++ {
		slice[i] = value
	}

	return slice
}

func Filter[T any](slice []T, predicate Predicate[T]) []T {
	result := make([]T, 0, len(slice))

	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}

	return result
}

func Find[T any](slice []T, predicate Predicate[T]) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func FindIndex[T any](slice []T, predicate Predicate[T]) int {
	for i, v := range slice {
		if predicate(v) {
			return i
		}
	}

	return -1
}

func FindLast[T any](slice []T, predicate Predicate[T]) (T, bool) {
	for i := len(slice) - 1; i >= 0; i-- {
		if predicate(slice[i]) {
			return slice[i], true
		}
	}

	var zero T

	return zero, false
}

func FindLastIndex[T any](slice []T, predicate Predicate[T]) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if predicate(slice[i]) {
			return i
		}
	}

	return -1
}

func Flat[T any](nestedSlice [][]T) []T {
	totalLen := 0
	for _, innerSlice := range nestedSlice {
		totalLen += len(innerSlice)
	}

	result := make([]T, 0, totalLen)

	for _, innerSlice := range nestedSlice {
		result = append(result, innerSlice...)
	}

	return result
}

func FlatMap[T any, U any](slice []T, transform func(T) []U) []U {
	var result []U

	for _, v := range slice {
		result = append(result, transform(v)...)
	}

	return result
}

func ForEach[T any](slice []T, consumer func(T)) {
	for _, v := range slice {
		consumer(v)
	}
}

func Includes[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}

	return false
}

func IndexOf[T comparable](slice []T, element T) int {
	for i, v := range slice {
		if v == element {
			return i
		}
	}

	return -1
}

func IsSlice(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func Join(slice []string, sep string) string {
	return strings.Join(slice, sep)
}

func LastIndexOf[T comparable](slice []T, element T, fromIndex int) int {
	length := len(slice)
	if fromIndex >= length {
		fromIndex = length - 1
	}

	if fromIndex < 0 {
		return -1
	}

	for i := fromIndex; i >= 0; i-- {
		if slice[i] == element {
			return i
		}
	}

	return -1
}

func Map[T any, U any](slice []T, transform func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = transform(v)
	}

	return result
}

func Pop[T any](slice []T) (T, []T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, slice, false
	}

	val := slice[len(slice)-1]
	return val, slice[:len(slice)-1], true
}

func Push[T any](slice []T, val T) []T {
	return append(slice, val)
}

func Reduce[T any, U any](slice []T, initial U, accumulator func(U, T) U) U {
	acc := initial
	for _, v := range slice {
		acc = accumulator(acc, v)
	}
	return acc
}

func ReduceRight[T any, U any](slice []T, initial U, accumulator func(U, T) U) U {
	acc := initial
	for i := len(slice) - 1; i >= 0; i-- {
		acc = accumulator(acc, slice[i])
	}

	return acc
}

func Reverse[T any](slice []T) []T {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}

	return slice
}

func Shift[T any](slice []T) (T, []T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, slice, false
	}
	val := slice[0]
	return val, slice[1:], true
}

func Select[T any](slice []T, start, end int) []T {
	length := len(slice)

	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start >= end {
		return []T{}
	}

	return slice[start:end]
}

func Some[T any](slice []T, predicate Predicate[T]) bool {
	for _, v := range slice {
		if predicate(v) {
			return true
		}
	}

	return false
}
func Sort[T cmp.Ordered](slice []T) []T {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})

	return slice
}

func SortFunc[T any](slice []T, less func(a, b T) bool) []T {
	sort.Slice(slice, func(i, j int) bool {
		return less(slice[i], slice[j])
	})

	return slice
}

func Splice[T any](slice []T, start, deleteCount int, items ...T) ([]T, []T) {
	length := len(slice)
	if start < 0 {
		start = length + start
	}
	if start < 0 {
		start = 0
	}
	if start > length {
		start = length
	}

	if deleteCount < 0 {
		deleteCount = 0
	}
	if start+deleteCount > length {
		deleteCount = length - start
	}

	removed := make([]T, deleteCount)
	copy(removed, slice[start:start+deleteCount])

	newSlice := make([]T, length-deleteCount+len(items))
	copy(newSlice, slice[:start])
	copy(newSlice[start:], items)
	copy(newSlice[start+len(items):], slice[start+deleteCount:])

	return removed, newSlice
}

func ToReversed[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	return Reverse(result)
}

func ToSorted[T cmp.Ordered](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	return Sort(result)
}

func ToSortedFunc[T any](slice []T, less func(a, b T) bool) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	return SortFunc(result, less)
}

func ToSpliced[T any](slice []T, start, deleteCount int, items ...T) []T {
	_, newSlice := Splice(slice, start, deleteCount, items...)
	return newSlice
}

func ToString[T any](slice []T) string {
	var sb strings.Builder
	sb.WriteString("[")

	for i, v := range slice {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v", v))
	}

	sb.WriteString("]")
	return sb.String()
}

func Unshift[T any](slice []T, elements ...T) []T {
	return append(elements, slice...)
}

func With[T any](slice []T, index int, value T) []T {
	length := len(slice)
	if index < 0 {
		index = length + index
	}
	if index < 0 || index >= length {
		return slice
	}

	result := make([]T, length)
	copy(result, slice)
	result[index] = value

	return result
}
