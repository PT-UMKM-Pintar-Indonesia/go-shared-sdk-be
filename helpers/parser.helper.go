package sdk_helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"
	"time"

	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	"github.com/google/uuid"
)

type parser struct{}

func NewParser() sdk_inf.IParser {
	return &parser{}
}

func (h *parser) ToString(v any) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	case []byte:
		return strings.TrimSpace(string(s))
	case int:
		return strconv.Itoa(s)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func (h *parser) ToInt(v any) (int, error) {
	s := h.ToString(v)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func (h *parser) ToFloat(v any) (float64, error) {
	return strconv.ParseFloat(h.ToString(v), 64)
}

func (h *parser) ToByte(v any) ([]byte, error) {
	return []byte(h.ToString(v)), nil
}

func (h *parser) Marshal(src any) ([]byte, error) {
	return json.Marshal(src)
}

func (h *parser) Unmarshal(src []byte, dest any) error {
	return json.Unmarshal(src, dest)
}

func (h *parser) Decode(src io.Reader, dest any) error {
	return json.NewDecoder(src).Decode(dest)
}

func (h *parser) Encode(src io.Writer, dest any) error {
	return json.NewEncoder(src).Encode(dest)
}

func (h *parser) FromUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func (h *parser) FromNullUUID(s string) (uuid.NullUUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.NullUUID{}, err
	}
	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

func (h *parser) DecimalToFloat(n int64) float64 {
	return float64(n) / 100.0
}

func (h *parser) HtmlFileToStr(filename string, data any) (string, error) {
	fullPath := filename
	if !strings.HasSuffix(filename, ".html") {
		fullPath += ".html"
	}

	tmp, err := template.ParseFiles(fullPath)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	if data != nil {
		if m, ok := data.(map[string]any); ok {
			if _, exists := m["year"]; !exists {
				m["year"] = time.Now().Year()
			}
		}
	}

	var buf bytes.Buffer
	if err := tmp.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	return buf.String(), nil
}
