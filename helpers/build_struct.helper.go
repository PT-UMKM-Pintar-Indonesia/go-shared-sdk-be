package sdk_helper

import (
	"fmt"
	"reflect"
	"sync"
)

type StructObject struct {
	val reflect.Value
	typ reflect.Type
}

var fieldCache sync.Map

func NewStructObject(v any) (*StructObject, error) {
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a pointer to a struct")
	}

	return &StructObject{
		val: val.Elem(),
		typ: val.Elem().Type(),
	}, nil
}

func (h *StructObject) getFieldIndex(name string) (int, bool) {
	cacheKey := fmt.Sprintf("%s:%s", h.typ.String(), name)

	if idx, ok := fieldCache.Load(cacheKey); ok {
		return idx.(int), true
	}

	field, ok := h.typ.FieldByName(name)
	if !ok {
		return -1, false
	}

	fieldCache.Store(cacheKey, field.Index[0])
	return field.Index[0], true
}

func (h *StructObject) Get(name string) (any, bool) {
	idx, ok := h.getFieldIndex(name)

	if !ok {
		return nil, false
	}

	return h.val.Field(idx).Interface(), true
}

func (h *StructObject) Set(name string, value any) error {
	idx, ok := h.getFieldIndex(name)
	if !ok {
		return fmt.Errorf("field %s not found", name)
	}

	fieldVal := h.val.Field(idx)
	if !fieldVal.CanSet() {
		return fmt.Errorf("field %s cannot be set", name)
	}

	val := reflect.ValueOf(value)
	if val.Type() != fieldVal.Type() {
		return fmt.Errorf("type mismatch for field %s: expected %s, got %s", name, fieldVal.Type(), val.Type())
	}

	fieldVal.Set(val)
	return nil
}

func (h *StructObject) Keys() []string {
	keys := make([]string, 0, h.typ.NumField())

	for i := 0; i < h.typ.NumField(); i++ {
		field := h.typ.Field(i)
		if field.PkgPath == "" {
			keys = append(keys, field.Name)
		}
	}

	return keys
}

func (h *StructObject) Assign(source any) error {
	srcVal := reflect.ValueOf(source)

	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}

	if srcVal.Kind() != reflect.Struct {
		return fmt.Errorf("source must be a struct")
	}

	srcTyp := srcVal.Type()
	for i := 0; i < srcTyp.NumField(); i++ {
		field := srcTyp.Field(i)

		if field.PkgPath != "" {
			continue
		}

		_ = h.Set(field.Name, srcVal.Field(i).Interface())
	}

	return nil
}
