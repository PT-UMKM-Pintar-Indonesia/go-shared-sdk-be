package sdk_helper

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type parser struct{}

func NewParser() sdk_inf.IParser {
	return parser{}
}

func (h parser) ToString(v any) string {
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func (h parser) ToInt(v any) (int, error) {
	parse, err := strconv.Atoi(h.ToString(v))
	if err != nil {
		return 0, nil
	}

	return parse, nil
}

func (h parser) ToFloat(v any) (float64, error) {
	parse, err := strconv.ParseFloat(h.ToString(v), 64)
	if err != nil {
		return 0, err
	}

	return parse, nil
}

func (h parser) ToByte(v any) ([]byte, error) {
	reader := strings.NewReader(h.ToString(v))
	data := &bytes.Buffer{}

	if _, err := reader.WriteTo(data); err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (h parser) Marshal(src any) ([]byte, error) {
	return json.Marshal(src)
}

func (h parser) Unmarshal(src []byte, dest any) error {
	decoder := json.NewDecoder(bytes.NewReader(src))

	for decoder.More() {
		if err := decoder.Decode(dest); err != nil {
			return err
		}
	}

	return nil
}

func (h parser) Decode(src io.Reader, dest any) error {
	decoder := json.NewDecoder(src)

	for decoder.More() {
		if err := decoder.Decode(dest); err != nil {
			return err
		}
	}

	return nil
}

func (h parser) Encode(src io.Writer, dest any) error {
	return json.NewEncoder(src).Encode(dest)
}

func (h parser) FromUUID(s string) uuid.UUID {
	fromId, err := uuid.FromBytes([]byte(s))
	if err != nil {
		logrus.Errorf("FromUUID: %v", err)
	}

	return fromId
}

func (h parser) FromNullUUID(s string) uuid.NullUUID {
	fromId, err := uuid.FromBytes([]byte(s))
	if err != nil {
		logrus.Errorf("FromNullUUID: %v", err)
	}
	return uuid.NullUUID{UUID: fromId, Valid: true}
}

func (h parser) DecimalToFloat(n int64) float64 {
	return float64(n) / 100
}

func (h parser) HtmlFileToStr(filename string, data any) (string, error) {
	tmp, err := template.ParseFiles(filename + ".html")
	if err != nil {
		return "", err
	}

	writer := new(bytes.Buffer)

	if data != nil && reflect.ValueOf(data).Kind() == reflect.Map {
		year := time.Now().Year()
		data.(map[string]any)["year"] = year
	}

	if err := tmp.Execute(writer, data); err != nil {
		return "", err
	}

	return writer.String(), nil
}
