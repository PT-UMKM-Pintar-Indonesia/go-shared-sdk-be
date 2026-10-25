package sdk_pkg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func captureDefaultLogrus(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer

	oldOut := logrus.StandardLogger().Out
	oldLevel := logrus.GetLevel()

	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.TraceLevel)

	t.Cleanup(func() {
		logrus.SetOutput(oldOut)
		logrus.SetLevel(oldLevel)
	})

	fn()

	return buf.String()
}

// A5 regression: Logrus panicked when Args were present but Msg was not a
// string (unchecked Msg.(string) assertion).
func TestLogrusNonStringMsgWithArgsNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Logrus panicked on non-string Msg: %v", r)
		}
	}()

	out := captureDefaultLogrus(t, func() {
		Logrus(sdk_cons.ERROR, map[string]any{"ctx": "value"}, "ignored-arg")
	})

	if !strings.Contains(out, "ctx") {
		t.Fatalf("expected structured fallback output, got %q", out)
	}
}

func TestLogrusFormatPath(t *testing.T) {
	out := captureDefaultLogrus(t, func() {
		Logrus(sdk_cons.INFO, "hello %s #%d", "world", 7)
	})

	if !strings.Contains(out, "hello world #7") {
		t.Fatalf("formatted output missing, got %q", out)
	}
}

func TestLogrusPlainMsgPath(t *testing.T) {
	out := captureDefaultLogrus(t, func() {
		Logrus(sdk_cons.WARN, "no args here")
	})

	if !strings.Contains(out, "no args here") {
		t.Fatalf("expected plain output, got %q", out)
	}
}

func TestLogrusPanicTypeStillPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("PANIC type must panic")
		}
	}()

	_ = captureDefaultLogrus(t, func() {
		Logrus(sdk_cons.PANIC, "boom %s", "hard")
	})
}

func TestLogrusConcurrent(t *testing.T) {
	// A4/A5 companion: no formatter races or assertion panics under
	// concurrency (run with -race in CI).
	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			Logrus(sdk_cons.INFO, "worker %d", i)
			Logrus(sdk_cons.ERROR, struct{ A int }{A: i}, "extra")
		}(i)
	}

	wg.Wait()
}

func TestLogrusArgsNormalisation(t *testing.T) {
	if got := logrusArgs([]any{1, "x"}); len(got) != 2 {
		t.Fatalf("slice must be spread, got %v", got)
	}

	if got := logrusArgs(42); len(got) != 1 || got[0] != 42 {
		t.Fatalf("scalar must be single arg, got %v", got)
	}

	if got := logrusArgs(nil); len(got) != 1 || got[0] != nil {
		t.Fatalf("nil must be passed through, got %v", got)
	}
}

func TestLogrusCustomLoggerSliceArgs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.log")

	out, err := LogrusCustomLogger(&sdk_dto.LogrusCustomLogger{
		FileName:      file,
		Type:          sdk_cons.INFO,
		CustomMessage: "count %d name %s",
		Args:          []any{3, "three"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	// B-fix regression: Args used to be passed as a single any value,
	// producing "count %!d(...)" garbage in the log line.
	if !strings.Contains(string(raw), "count 3 name three") {
		t.Fatalf("formatted args missing, file=%q", raw)
	}

	_ = out
}

func TestLogrusCustomLoggerScalarArgs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "scalar.log")

	_, err := LogrusCustomLogger(&sdk_dto.LogrusCustomLogger{
		FileName:      file,
		Type:          sdk_cons.PRINT,
		CustomMessage: "single %v",
		Args:          "scalar",
	})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(raw), "single scalar") {
		t.Fatalf("unexpected file content %q", raw)
	}
}

func TestLogrusCustomLoggerJSONFields(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "fields.log")

	_, err := LogrusCustomLogger(&sdk_dto.LogrusCustomLogger{
		FileName: file,
		Type:     sdk_cons.ERROR,
		Fields: logrus.Fields{
			"trace_id": "abc123",
		},
		Args:          "denied",
		CustomMessage: "user denied: %v",
	})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	body := string(raw)

	for _, want := range []string{"abc123", "user denied: denied"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %q", want, body)
		}
	}
}

func TestLogrusCustomLoggerCustomFields(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "custom.log")

	// CustomFields replaces the base Fields set (existing SDK semantics);
	// the call must succeed and emit the formatted message.
	_, err := LogrusCustomLogger(&sdk_dto.LogrusCustomLogger{
		FileName: file,
		Type:     sdk_cons.ERROR,
		CustomFields: &sdk_dto.LogrusCustomFields{
			Code: "AUTH_FAIL",
			Msg:  "denied",
		},
		Args:          "denied",
		CustomMessage: "user denied: %v",
	})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(raw), "user denied: denied") {
		t.Fatalf("unexpected file content %q", raw)
	}
}

func TestLogrusCustomLoggerFileError(t *testing.T) {
	_, err := LogrusCustomLogger(&sdk_dto.LogrusCustomLogger{
		FileName:      "/nonexistent-dir-xyz/out.log",
		Type:          sdk_cons.INFO,
		CustomMessage: "x %v",
		Args:          1,
	})

	if err == nil {
		t.Fatal("expected error writing to invalid path")
	}
}
