package sdk_helper

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func writeFileForTest(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}

func TestParserToStringToInt(t *testing.T) {
	p := NewParser()

	if p.ToString(nil) != "" {
		t.Fatal("nil -> empty")
	}

	if p.ToString("  padded  ") != "padded" {
		t.Fatal("string trim")
	}

	if p.ToString([]byte(" raw ")) != "raw" {
		t.Fatal("bytes trim")
	}

	if p.ToString(42) != "42" {
		t.Fatal("int")
	}

	n, err := p.ToInt(" 42 ")
	if err != nil || n != 42 {
		t.Fatalf("ToInt %d %v", n, err)
	}

	if n, err := p.ToInt(""); err != nil || n != 0 {
		t.Fatalf("empty should be 0, got %d %v", n, err)
	}

	if _, err := p.ToInt("abc"); err == nil {
		t.Fatal("expected error")
	}

	f, err := p.ToFloat("1.5")
	if err != nil || f != 1.5 {
		t.Fatalf("ToFloat %v %v", f, err)
	}

	b, err := p.ToByte("abc")
	if err != nil || string(b) != "abc" {
		t.Fatalf("ToByte %q %v", b, err)
	}
}

func TestParserJSON(t *testing.T) {
	p := NewParser()

	raw, err := p.Marshal(map[string]int{"a": 1})
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]int
	if err := p.Unmarshal(raw, &m); err != nil || m["a"] != 1 {
		t.Fatalf("unmarshal %v %v", m, err)
	}

	var enc bytes.Buffer
	if err := p.Encode(&enc, m); err != nil {
		t.Fatal(err)
	}

	var dec map[string]int
	if err := p.Decode(bytes.NewReader(enc.Bytes()), &dec); err != nil || dec["a"] != 1 {
		t.Fatalf("decode %v %v", dec, err)
	}

	if err := p.Unmarshal([]byte("{bad"), &m); err == nil {
		t.Fatal("expected json error")
	}
}

func TestParserUUID(t *testing.T) {
	p := NewParser()

	id, err := p.FromUUID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil || id.String() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("FromUUID %v %v", id, err)
	}

	if _, err := p.FromUUID("nope"); err == nil {
		t.Fatal("expected parse error")
	}

	nu, err := p.FromNullUUID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil || !nu.Valid {
		t.Fatalf("FromNullUUID %v %v", nu, err)
	}

	if _, err := p.FromNullUUID("nope"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParserDecimalToFloat(t *testing.T) {
	p := NewParser()

	if got := p.DecimalToFloat(1234); got != 12.34 {
		t.Fatalf("got %v", got)
	}

	if got := p.DecimalToFloat(0); got != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestParserHtmlFileToStr(t *testing.T) {
	p := NewParser()

	if _, err := p.HtmlFileToStr("does-not-exist", nil); err == nil {
		t.Fatal("expected template parse error for missing file")
	}

	dir := t.TempDir()
	file := dir + "/page.html"

	if err := writeFileForTest(file, "<p>Hello {{.Name}} {{.year}}</p>"); err != nil {
		t.Fatal(err)
	}

	out, err := p.HtmlFileToStr(file, map[string]any{"Name": "World"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if !strings.Contains(out, "Hello World") {
		t.Fatalf("unexpected output %q", out)
	}

	// .html suffix is appended automatically.
	if out2, err := p.HtmlFileToStr(dir+"/page", map[string]any{"Name": "Again"}); err != nil || !strings.Contains(out2, "Again") {
		t.Fatalf("suffix handling: %q %v", out2, err)
	}
}
