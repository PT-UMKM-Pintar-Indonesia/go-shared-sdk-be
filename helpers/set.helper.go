package sdk_helper

import "sync"

type (
	SetSafe[T comparable] struct {
		elements map[T]struct{}
		mu       sync.RWMutex
	}

	SetUnSafe[T comparable] struct {
		elements map[T]struct{}
	}
)

func NewSetSafe[T comparable](initialElements ...T) *SetSafe[T] {
	h := &SetSafe[T]{
		elements: make(map[T]struct{}, len(initialElements)),
	}

	h.mu.Lock()

	for _, el := range initialElements {
		h.elements[el] = struct{}{}
	}

	h.mu.Unlock()
	return h
}

func (h *SetSafe[T]) Add(element T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.elements[element] = struct{}{}
}

func (h *SetSafe[T]) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.elements = make(map[T]struct{})
}

func (h *SetSafe[T]) Delete(element T) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, exists := h.elements[element]
	if exists {
		delete(h.elements, element)
		return true
	}

	return false
}

func (h *SetSafe[T]) Has(element T) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.elements[element]
	return exists
}

func (h *SetSafe[T]) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.elements)
}

func (h *SetSafe[T]) Values() []T {
	h.mu.RLock()
	defer h.mu.RUnlock()

	values := make([]T, 0, len(h.elements))
	for el := range h.elements {
		values = append(values, el)
	}

	return values
}

func (h *SetSafe[T]) ForEach(callback func(element T)) {
	h.mu.RLock()

	elementsCopy := make([]T, 0, len(h.elements))
	for el := range h.elements {
		elementsCopy = append(elementsCopy, el)
	}
	h.mu.RUnlock()

	for _, el := range elementsCopy {
		callback(el)
	}
}

func (h *SetSafe[T]) Entries() []struct{ Value T } {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries := make([]struct{ Value T }, 0, len(h.elements))
	for el := range h.elements {
		entries = append(entries, struct{ Value T }{Value: el})
	}

	return entries
}

func (h *SetSafe[T]) Union(other *SetSafe[T]) *SetSafe[T] {
	result := NewSetSafe[T]()

	h.mu.RLock()
	for el := range h.elements {
		result.Add(el)
	}
	h.mu.RUnlock()

	other.mu.RLock()
	for el := range other.elements {
		result.Add(el)
	}
	other.mu.RUnlock()

	return result
}

func (h *SetSafe[T]) Intersection(other *SetSafe[T]) *SetSafe[T] {
	result := NewSetSafe[T]()

	if h.Size() < other.Size() {
		h.mu.RLock()
		other.mu.RLock()

		for el := range h.elements {
			if other.Has(el) {
				result.Add(el)
			}
		}

		other.mu.RUnlock()
		h.mu.RUnlock()
	} else {
		other.mu.RLock()
		h.mu.RLock()

		for el := range other.elements {
			if h.Has(el) {
				result.Add(el)
			}
		}

		h.mu.RUnlock()
		other.mu.RUnlock()
	}

	return result
}

func (h *SetSafe[T]) Difference(other *SetSafe[T]) *SetSafe[T] {
	result := NewSetSafe[T]()

	h.mu.RLock()
	other.mu.RLock()

	for el := range h.elements {
		if !other.Has(el) {
			result.Add(el)
		}
	}

	other.mu.RUnlock()
	h.mu.RUnlock()

	return result
}

func (h *SetSafe[T]) SymmetricDifference(other *SetSafe[T]) *SetSafe[T] {
	result := NewSetSafe[T]()

	h.mu.RLock()
	other.mu.RLock()

	for el := range h.elements {
		if !other.Has(el) {
			result.Add(el)
		}
	}
	for el := range other.elements {
		if !h.Has(el) {
			result.Add(el)
		}
	}

	other.mu.RUnlock()
	h.mu.RUnlock()

	return result
}

func (h *SetSafe[T]) IsSubsetOf(other *SetSafe[T]) bool {
	h.mu.RLock()
	other.mu.RLock()
	defer h.mu.RUnlock()
	defer other.mu.RUnlock()

	if h.Size() > other.Size() {
		return false
	}

	for el := range h.elements {
		if !other.Has(el) {
			return false
		}
	}

	return true
}

func (h *SetSafe[T]) IsSupersetOf(other *SetSafe[T]) bool {
	return other.IsSubsetOf(h)
}

func (h *SetSafe[T]) IsDisjointFrom(other *SetSafe[T]) bool {
	h.mu.RLock()
	other.mu.RLock()
	defer h.mu.RUnlock()
	defer other.mu.RUnlock()

	if h.Size() < other.Size() {
		for el := range h.elements {
			if other.Has(el) {
				return false
			}
		}
	} else {
		for el := range other.elements {
			if h.Has(el) {
				return false
			}
		}
	}

	return true
}

// TODO: set Unsafe

func NewSetUnSafe[T comparable](initialElements ...T) *SetUnSafe[T] {
	h := &SetUnSafe[T]{
		elements: make(map[T]struct{}, len(initialElements)),
	}

	for _, el := range initialElements {
		h.Add(el)
	}

	return h
}

func (h *SetUnSafe[T]) Add(element T) {
	h.elements[element] = struct{}{}
}

func (h *SetUnSafe[T]) Clear() {
	h.elements = make(map[T]struct{})
}

func (h *SetUnSafe[T]) Delete(element T) bool {
	_, exists := h.elements[element]

	if exists {
		delete(h.elements, element)
		return true
	}

	return false
}

func (h *SetUnSafe[T]) Has(element T) bool {
	_, exists := h.elements[element]
	return exists
}

func (h *SetUnSafe[T]) Size() int {
	return len(h.elements)
}

func (h *SetUnSafe[T]) Values() []T {
	values := make([]T, 0, h.Size())

	for el := range h.elements {
		values = append(values, el)
	}

	return values
}

func (h *SetUnSafe[T]) ForEach(callback func(element T)) {
	for el := range h.elements {
		callback(el)
	}
}

func (h *SetUnSafe[T]) Entries() []struct{ Value T } {
	entries := make([]struct{ Value T }, 0, h.Size())

	for el := range h.elements {
		entries = append(entries, struct{ Value T }{Value: el})
	}

	return entries
}

func (h *SetUnSafe[T]) Union(other *SetUnSafe[T]) *SetUnSafe[T] {
	result := NewSetUnSafe[T]()

	h.ForEach(func(el T) {
		result.Add(el)
	})

	other.ForEach(func(el T) {
		result.Add(el)
	})

	return result
}

func (h *SetUnSafe[T]) Intersection(other *SetUnSafe[T]) *SetUnSafe[T] {
	result := NewSetUnSafe[T]()

	if h.Size() < other.Size() {
		h.ForEach(func(el T) {
			if other.Has(el) {
				result.Add(el)
			}
		})
	} else {
		other.ForEach(func(el T) {
			if h.Has(el) {
				result.Add(el)
			}
		})
	}

	return result
}

func (h *SetUnSafe[T]) Difference(other *SetUnSafe[T]) *SetUnSafe[T] {
	result := NewSetUnSafe[T]()
	h.ForEach(func(el T) {
		if !other.Has(el) {
			result.Add(el)
		}
	})

	return result
}

func (h *SetUnSafe[T]) SymmetricDifference(other *SetUnSafe[T]) *SetUnSafe[T] {
	result := NewSetUnSafe[T]()

	h.ForEach(func(el T) {
		if !other.Has(el) {
			result.Add(el)
		}
	})

	other.ForEach(func(el T) {
		if !h.Has(el) {
			result.Add(el)
		}
	})

	return result
}

func (h *SetUnSafe[T]) IsSubsetOf(other *SetUnSafe[T]) bool {
	if h.Size() > other.Size() {
		return false
	}
	isSubset := true
	h.ForEach(func(el T) {
		if !other.Has(el) {
			isSubset = false
			return
		}
	})

	return isSubset
}

func (h *SetUnSafe[T]) IsSupersetOf(other *SetUnSafe[T]) bool {
	return other.IsSubsetOf(h)
}

func (h *SetUnSafe[T]) IsDisjointFrom(other *SetUnSafe[T]) bool {
	if h.Size() < other.Size() {
		isDisjoint := true
		h.ForEach(func(el T) {
			if other.Has(el) {
				isDisjoint = false

				return
			}
		})

		return isDisjoint
	} else {
		isDisjoint := true
		other.ForEach(func(el T) {
			if h.Has(el) {
				isDisjoint = false
				return
			}
		})

		return isDisjoint
	}
}
