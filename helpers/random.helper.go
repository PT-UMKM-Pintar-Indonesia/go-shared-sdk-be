package sdk_helper

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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

	var sb strings.Builder
	sb.Grow(length)

	chLen := len(charset)
	mask := uint64(1)<<6 - 1
	if chLen > 64 {
		mask = uint64(1)<<7 - 1
	}

	buffer := make([]byte, length+(length/4))
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	idx := 0
	for sb.Len() < length {
		if idx >= len(buffer) {
			buffer = make([]byte, length-sb.Len())
			if _, err := rand.Read(buffer); err != nil {
				return "", err
			}
			idx = 0
		}

		val := int(uint64(buffer[idx]) & mask)
		if val < chLen {
			sb.WriteByte(charset[val])
		}
		idx++
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
