package sdk_helper

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

type (
	ICipher interface {
		Base64Encode(plainText string) string
		Base64Decode(cipherText string) (string, error)
		EncodeRotation(plainText string) string
		DecodeRotation(cipherText string) (string, error)
		CaesarEncrypt(plainText string, rotation int) string
		CaesarDecrypt(cipherText string, rotation int) string
		RotateNumber(plainNumber string, shift int) string
		DerotateNumber(cipherNumber string, shift int) string
	}

	cipher struct{}
)

func NewCipher() ICipher {
	return &cipher{}
}

func (h *cipher) Base64Encode(plainText string) string {
	return base64.StdEncoding.EncodeToString([]byte(plainText))
}

func (h *cipher) Base64Decode(cipherText string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}
	return string(decoded), nil
}

func (h *cipher) EncodeRotation(plainText string) string {
	timestamp := time.Now().Format("20060102150405")
	payload := fmt.Sprintf("%s:%s", plainText, timestamp)

	b64 := h.Base64Encode(payload)
	n := len(b64)
	if n == 0 {
		return ""
	}

	shift := 32 % n
	rotated := b64[shift:] + b64[:shift]

	return h.CaesarEncrypt(rotated, 26)
}

func (h *cipher) DecodeRotation(cipherText string) (string, error) {
	decrypted := h.CaesarDecrypt(cipherText, 26)
	n := len(decrypted)
	if n == 0 {
		return "", errors.New("empty cipher text")
	}

	shift := 32 % n
	originalB64 := decrypted[n-shift:] + decrypted[:n-shift]

	return h.Base64Decode(originalB64)
}

func (h *cipher) CaesarEncrypt(plainText string, rotation int) string {
	return h.applyCaesar(plainText, rotation)
}

func (h *cipher) CaesarDecrypt(cipherText string, rotation int) string {
	return h.applyCaesar(cipherText, -rotation)
}

func (h *cipher) applyCaesar(input string, rotation int) string {
	var builder strings.Builder
	builder.Grow(len(input))
	shift := rune((rotation%26 + 26) % 26)
	if shift == 0 {
		return input
	}

	for _, char := range input {
		switch {
		case char >= 'A' && char <= 'Z':
			builder.WriteRune((char-'A'+shift)%26 + 'A')
		case char >= 'a' && char <= 'z':
			builder.WriteRune((char-'a'+shift)%26 + 'a')
		default:
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func (h *cipher) RotateNumber(plainNumber string, rotation int) string {
	shift := rune((rotation%10 + 10) % 10)
	if shift == 0 {
		return plainNumber
	}

	var builder strings.Builder
	builder.Grow(len(plainNumber))

	for _, r := range plainNumber {
		if r >= '0' && r <= '9' {
			builder.WriteRune((r-'0'+shift)%10 + '0')
		} else {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func (h *cipher) DerotateNumber(cipherNumber string, rotation int) string {
	return h.RotateNumber(cipherNumber, -rotation)
}
