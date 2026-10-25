package sdk_helper

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	mr "math/rand"
	"strings"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

const (
	charsetAlpha        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetNumeric      = "0123456789"
	charsetAlphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type random struct{}

func NewRandom() sdk_inf.IRandom {
	return &random{}
}

func (h *random) generate(length int, charset string) (string, error) {
	if length <= 0 {
		return "", nil
	}

	if len(charset) == 0 {
		return "", errors.New("charset cannot be empty")
	}

	max := big.NewInt(int64(len(charset)))
	var sb strings.Builder
	sb.Grow(length)

	for sb.Len() < length {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		sb.WriteByte(charset[n.Int64()])
	}

	return sb.String(), nil
}

func (h *random) AlphaCharacters(length int) string {
	res, err := h.generate(length, charsetAlpha)
	if err != nil {
		return sdk_cons.EMPTY
	}
	return res
}

func (h *random) Numeric(length int) string {
	if length <= 0 {
		return sdk_cons.EMPTY
	}

	firstDigitCharset := "123456789"
	first, err := h.generate(1, firstDigitCharset)
	if err != nil {
		return sdk_cons.EMPTY
	}

	if length == 1 {
		return first
	}

	rest, err := h.generate(length-1, charsetNumeric)
	if err != nil {
		return sdk_cons.EMPTY
	}

	return first + rest
}

func (h *random) Alphanumeric(length int) (string, error) {
	if length <= 0 {
		return sdk_cons.EMPTY, errors.New("length must be greater than 0")
	}

	return h.generate(length, charsetAlphanumeric)
}

func (h *random) Secure(length int, charset string) (string, error) {
	if charset == "" {
		return sdk_cons.EMPTY, errors.New("charset cannot be empty")
	}

	return h.generate(length, charset)
}

func (h *random) Hex(length int) (string, error) {
	if length <= 0 {
		return sdk_cons.EMPTY, errors.New("length must be greater than 0")
	}

	byteLen := (length + 1) / 2
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	res := hex.EncodeToString(b)
	return res[:length], nil
}

func RandomNumericStr(length int) string {
	var builder strings.Builder
	builder.Grow(length)

	const digits = "0123456789"
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))

		if err != nil {
			return sdk_cons.EMPTY
		}

		builder.WriteByte(digits[num.Int64()])
	}

	return builder.String()
}

func RandomItem[T any](slice []T) T {
	var zeroValue T

	if len(slice) == 0 {
		return zeroValue
	}

	index := mr.Intn(len(slice))
	return slice[index]
}
