package sdk_helper

import (
	"fmt"
	"strings"
	"sync"
)

type (
	MapSafe[K comparable, V any] struct {
		elements map[K]V
		mu       sync.RWMutex
	}

	MapUnSafe[K comparable, V any] struct {
		elements map[K]V
	}
)

func NewMapSafe[K comparable, V any](initialPairs ...struct {
	Key   K
	Value V
}) *MapSafe[K, V] {
	h := &MapSafe[K, V]{
		elements: make(map[K]V, len(initialPairs)),
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, pair := range initialPairs {
		h.elements[pair.Key] = pair.Value
	}

	return h
}

func (h *MapSafe[K, V]) Set(key K, value V) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.elements[key] = value
}

func (h *MapSafe[K, V]) Get(key K) (V, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	val, ok := h.elements[key]
	return val, ok
}
func (h *MapSafe[K, V]) Has(key K) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.elements[key]
	return ok
}

func (h *MapSafe[K, V]) Delete(key K) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, exists := h.elements[key]
	if exists {
		delete(h.elements, key)
		return true
	}

	return false
}

func (h *MapSafe[K, V]) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.elements = make(map[K]V)
}

func (h *MapSafe[K, V]) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.elements)
}

func (h *MapSafe[K, V]) Keys() []K {
	h.mu.RLock()
	defer h.mu.RUnlock()

	keys := make([]K, 0, len(h.elements))
	for k := range h.elements {
		keys = append(keys, k)
	}

	return keys
}

func (h *MapSafe[K, V]) Values() []V {
	h.mu.RLock()
	defer h.mu.RUnlock()

	values := make([]V, 0, len(h.elements))
	for _, v := range h.elements {
		values = append(values, v)
	}

	return values
}

func (h *MapSafe[K, V]) Entries() []struct {
	Key   K
	Value V
} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	entries := make([]struct {
		Key   K
		Value V
	}, 0, len(h.elements))

	for k, v := range h.elements {
		entries = append(entries, struct {
			Key   K
			Value V
		}{Key: k, Value: v})
	}

	return entries
}

func (h *MapSafe[K, V]) ForEach(callback func(key K, value V)) {
	h.mu.RLock()

	elementsCopy := make(map[K]V, len(h.elements))
	for k, v := range h.elements {
		elementsCopy[k] = v
	}

	h.mu.RUnlock()
	for k, v := range elementsCopy {
		callback(k, v)
	}
}

func (h *MapSafe[K, V]) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var sb strings.Builder

	sb.WriteString("Map{")
	first := true

	for k, v := range h.elements {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v: %v", k, v))
		first = false
	}

	sb.WriteString("}")
	return sb.String()
}

func (h *MapSafe[K, V]) ValueOf() *MapSafe[K, V] {
	return h
}

// TODO: map UnSafe

func NewMapUnSafe[K comparable, V any](initialPairs ...struct {
	Key   K
	Value V
}) *MapUnSafe[K, V] {
	m := &MapUnSafe[K, V]{
		elements: make(map[K]V, len(initialPairs)),
	}

	for _, pair := range initialPairs {
		m.Set(pair.Key, pair.Value)
	}

	return m
}

func (m *MapUnSafe[K, V]) Set(key K, value V) {
	m.elements[key] = value
}

func (m *MapUnSafe[K, V]) Get(key K) (V, bool) {
	val, ok := m.elements[key]
	return val, ok
}

func (m *MapUnSafe[K, V]) Has(key K) bool {
	_, ok := m.elements[key]
	return ok
}

func (m *MapUnSafe[K, V]) Delete(key K) bool {
	_, exists := m.elements[key]
	if exists {
		delete(m.elements, key)
		return true
	}
	return false
}

func (m *MapUnSafe[K, V]) Clear() {
	m.elements = make(map[K]V)
}

func (m *MapUnSafe[K, V]) Size() int {
	return len(m.elements)
}

func (m *MapUnSafe[K, V]) Keys() []K {
	keys := make([]K, 0, m.Size())

	for k := range m.elements {
		keys = append(keys, k)
	}

	return keys
}

func (m *MapUnSafe[K, V]) Values() []V {
	values := make([]V, 0, m.Size())

	for _, v := range m.elements {
		values = append(values, v)
	}

	return values
}

func (m *MapUnSafe[K, V]) Entries() []struct {
	Key   K
	Value V
} {
	entries := make([]struct {
		Key   K
		Value V
	}, 0, m.Size())

	for k, v := range m.elements {
		entries = append(entries, struct {
			Key   K
			Value V
		}{Key: k, Value: v})
	}

	return entries
}

func (m *MapUnSafe[K, V]) ForEach(callback func(key K, value V)) {
	for k, v := range m.elements {
		callback(k, v)
	}
}

func (m *MapUnSafe[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("Map{")
	first := true

	for k, v := range m.elements {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v: %v", k, v))
		first = false
	}

	sb.WriteString("}")
	return sb.String()
}

func (m *MapUnSafe[K, V]) ValueOf() *MapUnSafe[K, V] {
	return m
}
