package sdk_helper

import "testing"

func TestEqualOps(t *testing.T) {
	if !Equal(1, 1) || Equal(1, 2) {
		t.Fatal("Equal")
	}

	if !NotEqual(1, 2) || NotEqual(1, 1) {
		t.Fatal("NotEqual")
	}
}

func TestGetTypeOfStringJSStyle(t *testing.T) {
	var nilIface any

	tests := []struct {
		name string
		val  any
		want string
	}{
		{"nil", nil, "null"},
		{"nil iface", nilIface, "null"},
		{"int", 1, "number"},
		{"float", 1.5, "number"},
		{"uint", uint8(2), "number"},
		{"string", "s", "string"},
		{"bool", true, "boolean"},
		{"slice", []int{}, "object"},
		{"map", map[string]int{}, "object"},
		{"struct", struct{}{}, "object"},
		{"ptr", new(int), "object"},
		{"func", func() {}, "function"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetTypeOfStringJSStyle(tc.val); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}

	if !IsTypeOf("x", "string") || IsTypeOf("x", "number") {
		t.Fatal("IsTypeOf")
	}
}

func TestIsPointerAndIsNil(t *testing.T) {
	if !IsPointer(new(int)) || IsPointer(42) || IsPointer(nil) {
		t.Fatal("IsPointer")
	}

	if !IsNil(nil) {
		t.Fatal("nil must be nil")
	}

	var p *int
	if !IsNil(p) {
		t.Fatal("typed nil pointer must be nil")
	}

	var m map[string]int
	if !IsNil(m) {
		t.Fatal("nil map must be nil")
	}

	if IsNil(42) || IsNil("") {
		t.Fatal("scalars are never nil")
	}

	if IsNil(new(int)) {
		t.Fatal("allocated pointer is not nil")
	}
}
