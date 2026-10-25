package sdk_helper

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type envDest struct {
	Name string `env:"SDK_TEST_NAME" mapstructure:"name"`
	Age  int    `env:"SDK_TEST_AGE" mapstructure:"age"`
}

// A4 regression: EnvToStruct used the viper package-level singleton, which
// leaks configuration between concurrent callers and is racy. Run under
// -race to validate the fix.
func TestEnvToStructConcurrent(t *testing.T) {
	t.Setenv("GO_ENV", "development")
	t.Setenv("SDK_TEST_NAME", "from-env")
	t.Setenv("SDK_TEST_AGE", "42")

	tr := NewTransform()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var dest envDest
			// No config file exists in the temp dir, so this must fall
			// back to env.Parse and still fill the struct.
			if err := tr.EnvToStruct("nonexistent-config", t.TempDir(), "json", &dest); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if dest.Name != "from-env" || dest.Age != 42 {
				t.Errorf("unexpected dest: %+v", dest)
			}
		}()
	}

	wg.Wait()
}

func TestEnvToStructProductionUsesEnvOnly(t *testing.T) {
	t.Setenv("GO_ENV", "production")
	t.Setenv("SDK_TEST_NAME", "prod")

	var dest envDest
	if err := NewTransform().EnvToStruct("whatever", ".", "json", &dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dest.Name != "prod" {
		t.Fatalf("expected env-only parse in production, got %+v", dest)
	}
}

func TestEnvToStructReadsConfigFile(t *testing.T) {
	t.Setenv("GO_ENV", "development")

	dir := t.TempDir()
	cfg := map[string]any{"name": "from-file", "age": 7}
	raw, _ := json.Marshal(cfg)

	if err := os.WriteFile(filepath.Join(dir, "app.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	var dest envDest
	if err := NewTransform().EnvToStruct("app", dir, "json", &dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dest.Name != "from-file" || dest.Age != 7 {
		t.Fatalf("expected config file values, got %+v", dest)
	}
}

func TestSrcToDest(t *testing.T) {
	tr := NewTransform()

	if err := tr.SrcToDest(nil, nil); err == nil {
		t.Fatal("expected error for nil src/dest")
	}

	type srcT struct {
		A string
		B int
	}
	type dstT struct {
		A string
		B int
	}

	var d dstT
	if err := tr.SrcToDest(srcT{A: "x", B: 3}, &d); err != nil {
		t.Fatalf("struct copy failed: %v", err)
	}

	if d != (dstT{A: "x", B: 3}) {
		t.Fatalf("unexpected dest %+v", d)
	}

	// Map source goes through JSON.
	var m map[string]any
	if err := tr.SrcToDest(map[string]any{"k": "v"}, &m); err != nil {
		t.Fatalf("map copy failed: %v", err)
	}

	if m["k"] != "v" {
		t.Fatalf("unexpected map %+v", m)
	}
}

func TestCtxToStruct(t *testing.T) {
	type payload struct {
		ID string
	}

	ctx := context.WithValue(context.Background(), "payload", payload{ID: "42"})

	var dest payload
	if err := NewTransform().CtxToStruct(ctx, "payload", &dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dest.ID != "42" {
		t.Fatalf("unexpected dest %+v", dest)
	}

	if err := NewTransform().CtxToStruct(ctx, "missing", &dest); err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestMapToStructAndBack(t *testing.T) {
	type item struct {
		Name string `json:"name"`
		Qty  int    `json:"qty"`
	}

	tr := NewTransform()

	var got item
	if err := tr.MapToStruct(map[string]any{"name": "apple", "qty": "3"}, &got); err != nil {
		t.Fatalf("weak typing decode failed: %v", err)
	}

	if got.Name != "apple" || got.Qty != 3 {
		t.Fatalf("unexpected %+v", got)
	}

	m, err := tr.StructToMap(got)
	if err != nil {
		t.Fatalf("struct to map failed: %v", err)
	}

	if m["name"] != "apple" {
		t.Fatalf("unexpected map %v", m)
	}
}

func TestConvert(t *testing.T) {
	tr := NewTransform()

	var s string
	if err := tr.Convert(42, &s); err != nil || s != "42" {
		t.Fatalf("int->string: %q %v", s, err)
	}

	var i int
	if err := tr.Convert("123", &i); err != nil || i != 123 {
		t.Fatalf("string->int: %d %v", i, err)
	}

	var f float64
	if err := tr.Convert(7, &f); err != nil || f != 7 {
		t.Fatalf("int->float: %v %v", f, err)
	}

	var b bool
	if err := tr.Convert("true", &b); err != nil || !b {
		t.Fatalf("string->bool: %v %v", b, err)
	}

	var tm time.Time
	if err := tr.Convert("2024-01-02", &tm); err != nil || tm.Year() != 2024 {
		t.Fatalf("string->time: %v %v", tm, err)
	}

	// nil src zeroes the destination.
	zero := 5
	if err := tr.Convert(nil, &zero); err != nil || zero != 0 {
		t.Fatalf("nil zeroing: %d %v", zero, err)
	}

	if err := tr.Convert(nil, 5); err == nil {
		t.Fatal("expected error for non-pointer dest")
	}
}

func TestUUIDHelpers(t *testing.T) {
	tr := NewTransform()

	const dashed = "550e8400-e29b-41d4-a716-446655440000"
	const bare = "550e8400e29b41d4a716446655440000"

	if got := tr.DecodeUUID(bare); got != dashed {
		t.Fatalf("DecodeUUID got %q", got)
	}

	if got := tr.EncodeUUID(dashed); got != bare {
		t.Fatalf("EncodeUUID got %q", got)
	}

	if got := tr.DecodeUUID("short"); got != "" {
		t.Fatalf("DecodeUUID invalid input should be empty, got %q", got)
	}

	if got := tr.EncodeUUID("short"); got != "" {
		t.Fatalf("EncodeUUID invalid input should be empty, got %q", got)
	}

	if got := tr.EncodeUUID("550e8400-zzzz-41d4-a716-446655440000"); got != "" {
		t.Fatalf("EncodeUUID non hex should be empty, got %q", got)
	}
}

func TestBodyToRaw(t *testing.T) {
	tr := NewTransform()

	if got := tr.BodyToRaw(nil); !strings.Contains(got, "error") {
		t.Fatalf("expected error json, got %q", got)
	}

	raw := tr.BodyToRaw(map[string]any{"a": 1})
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("invalid json %q: %v", raw, err)
	}
}
