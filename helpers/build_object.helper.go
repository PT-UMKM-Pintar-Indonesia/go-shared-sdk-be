package sdk_helper

import (
	"fmt"
	"strings"
	"sync"
)

type (
	MapObjectSafe[K comparable, V any] struct {
		mu   sync.RWMutex
		data map[K]V
	}

	MapObjectUnSafe[K comparable, V any] struct {
		data map[K]V
	}
)

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}

	return values
}

// TODO: map object safe use package

func NewMapObjectSafe[K comparable, V any](initial map[K]V) *MapObjectSafe[K, V] {
	data := make(map[K]V, len(initial))
	for k, v := range initial {
		data[k] = v
	}

	return &MapObjectSafe[K, V]{data: data}
}

func (h *MapObjectSafe[K, V]) Assign(source map[K]V) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for k, v := range source {
		h.data[k] = v
	}
}

func (h *MapObjectSafe[K, V]) Set(key K, value V) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data[key] = value
}

func (h *MapObjectSafe[K, V]) Get(key K) (V, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	v, ok := h.data[key]
	return v, ok
}

func (h *MapObjectSafe[K, V]) Delete(key K) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.data, key)
}

func (h *MapObjectSafe[K, V]) Has(key K) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.data[key]
	return ok
}

func (h *MapObjectSafe[K, V]) Keys() []K {
	h.mu.RLock()
	defer h.mu.RUnlock()
	keys := make([]K, 0, len(h.data))

	for k := range h.data {
		keys = append(keys, k)
	}

	return keys
}

func (h *MapObjectSafe[K, V]) Values() []V {
	h.mu.RLock()
	defer h.mu.RUnlock()
	values := make([]V, 0, len(h.data))

	for _, v := range h.data {
		values = append(values, v)
	}

	return values
}

func (h *MapObjectSafe[K, V]) Entries() []struct {
	Key   K
	Value V
} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	entries := make([]struct {
		Key   K
		Value V
	}, 0, len(h.data))
	for k, v := range h.data {
		entries = append(entries, struct {
			Key   K
			Value V
		}{Key: k, Value: v})
	}
	return entries
}

func (h *MapObjectSafe[K, V]) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var sb strings.Builder

	sb.WriteString("MapObjectSafe{")
	first := true

	for k, v := range h.data {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v: %v", k, v))
		first = false
	}

	sb.WriteString("}")
	return sb.String()
}

// TODO: map object unsafe use package

func NewMapObjectUnSafe[K comparable, V any](initial map[K]V) *MapObjectUnSafe[K, V] {
	data := make(map[K]V, len(initial))
	for k, v := range initial {
		data[k] = v
	}

	return &MapObjectUnSafe[K, V]{data: data}
}

func (h *MapObjectUnSafe[K, V]) Assign(source map[K]V) {
	for k, v := range source {
		h.data[k] = v
	}
}

func (h *MapObjectUnSafe[K, V]) Set(key K, value V) {
	h.data[key] = value
}

func (h *MapObjectUnSafe[K, V]) Get(key K) (V, bool) {
	v, ok := h.data[key]
	return v, ok
}

func (h *MapObjectUnSafe[K, V]) Delete(key K) {
	delete(h.data, key)
}

func (h *MapObjectUnSafe[K, V]) Has(key K) bool {
	_, ok := h.data[key]
	return ok
}

func (h *MapObjectUnSafe[K, V]) Keys() []K {
	keys := make([]K, 0, len(h.data))

	for k := range h.data {
		keys = append(keys, k)
	}

	return keys
}

func (h *MapObjectUnSafe[K, V]) Values() []V {
	values := make([]V, 0, len(h.data))

	for _, v := range h.data {
		values = append(values, v)
	}

	return values
}

func (h *MapObjectUnSafe[K, V]) Entries() []struct {
	Key   K
	Value V
} {

	entries := make([]struct {
		Key   K
		Value V
	}, 0, len(h.data))
	for k, v := range h.data {
		entries = append(entries, struct {
			Key   K
			Value V
		}{Key: k, Value: v})
	}
	return entries
}

func (h *MapObjectUnSafe[K, V]) String() string {
	var sb strings.Builder

	sb.WriteString("MapObjectUnSafe{")
	first := true

	for k, v := range h.data {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v: %v", k, v))
		first = false
	}

	sb.WriteString("}")
	return sb.String()
}
