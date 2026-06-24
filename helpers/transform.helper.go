package sdk_helper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-viper/mapstructure/v2"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"

	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

type transform struct{}

func NewTransform() sdk_inf.ITransform {
	return &transform{}
}

func (h *transform) SrcToDest(src, dest any) error {
	if src == nil || dest == nil {
		return errors.New("source or destination cannot be nil")
	}

	return copier.Copy(dest, src)
}

func (h *transform) CtxToStruct(ctx context.Context, key string, dest any) error {
	src := ctx.Value(key)
	if src == nil {
		return errors.New("key not found in context")
	}

	return copier.Copy(dest, src)
}

func (h *transform) EnvToStruct(name, path, ext string, dest any) error {
	if os.Getenv("GO_ENV") == "production" {
		return env.Parse(dest)
	}

	viper.SetConfigName(name)
	viper.SetConfigType(ext)
	viper.AddConfigPath(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return env.Parse(dest)
	}

	return viper.Unmarshal(dest)
}

func (h *transform) MapToStruct(src map[string]any, dest any) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           dest,
		WeaklyTypedInput: true,
		TagName:          "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(src); err != nil {
		return fmt.Errorf("failed to decode map to struct: %w", err)
	}

	return nil
}

func (h *transform) StructToMap(src any) (map[string]any, error) {
	var output map[string]any

	config := &mapstructure.DecoderConfig{
		Result:  &output,
		TagName: "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(src); err != nil {
		return nil, fmt.Errorf("failed to decode struct to map: %w", err)
	}

	return output, nil
}

func (h *transform) Convert(src any, dest any) error {
	if src == nil {
		outVal := reflect.ValueOf(dest)

		if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
			return fmt.Errorf("destination must be a non-nil pointer")
		}

		outVal.Elem().Set(reflect.Zero(outVal.Elem().Type()))
		return nil
	}

	switch out := dest.(type) {
	case *string:
		*out = fmt.Sprintf("%v", src)
		return nil

	case *int:
		if i, ok := src.(int); ok {
			*out = i
			return nil
		}
		if f, ok := src.(float64); ok {
			*out = int(f)
			return nil
		}
		if s, ok := src.(string); ok {
			if i, err := strconv.Atoi(s); err == nil {
				*out = i
				return nil
			}
		}

	case *float64:
		if f, ok := src.(float64); ok {
			*out = f
			return nil
		}

		if i, ok := src.(int); ok {
			*out = float64(i)
			return nil
		}

		if s, ok := src.(string); ok {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				*out = f
				return nil
			}
		}

	case *bool:
		if b, ok := src.(bool); ok {
			*out = b
			return nil
		}

		if s, ok := src.(string); ok {
			if b, err := strconv.ParseBool(s); err == nil {
				*out = b
				return nil
			}
		}

	case *time.Time:
		if t, ok := src.(time.Time); ok {
			*out = t
			return nil
		}
		if s, ok := src.(string); ok {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				*out = t
				return nil
			}

			if t, err := time.Parse("2006-01-02", s); err == nil {
				*out = t
				return nil
			}
		}
	}

	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           dest,
		WeaklyTypedInput: true,
		TagName:          "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("failed to create mapstructure decoder: %w", err)
	}

	if err := decoder.Decode(src); err != nil {
		return fmt.Errorf("failed to convert using mapstructure: %w", err)
	}

	return nil
}
