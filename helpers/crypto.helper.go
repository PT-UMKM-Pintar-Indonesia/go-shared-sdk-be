package sdk_helper

import (
	"crypto/aes"
	c "crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"

	"golang.org/x/crypto/scrypt"

	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

const (
	scryptN   = 16384
	scryptR   = 8
	scryptP   = 1
	keyLen    = 32
	saltLen   = 16
	minKeyLen = 16
)

type cryptoHelper struct{}

func NewCrypto() sdk_inf.ICrypto {
	return &cryptoHelper{}
}

func (h *cryptoHelper) deriveKey(secret, salt []byte) ([]byte, error) {
	return scrypt.Key(secret, salt, scryptN, scryptR, scryptP, keyLen)
}

func (h *cryptoHelper) AES256Encrypt(secretKey, plainText string) (string, error) {
	if len(secretKey) < minKeyLen {
		return "", fmt.Errorf("security error: secret key too short (min %d)", minKeyLen)
	}

	salt, err := h.GenerateRandomBytes(saltLen)
	if err != nil {
		return "", err
	}

	key, err := h.deriveKey([]byte(secretKey), salt)
	if err != nil {
		return "", fmt.Errorf("kdf failed: %w", err)
	}

	block, _ := aes.NewCipher(key)
	gcm, _ := c.NewGCM(block)

	nonce, err := h.GenerateRandomBytes(gcm.NonceSize())
	if err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	result := slices.Concat(salt, sealed)

	return hex.EncodeToString(result), nil
}

func (h *cryptoHelper) AES256Decrypt(secretKey, cipherTextHex string) (string, error) {
	data, err := hex.DecodeString(cipherTextHex)
	if err != nil || len(data) < saltLen+12 { // 16 (salt) + 12 (nonce)
		return "", errors.New("invalid ciphertext format")
	}

	salt, encryptedPayload := data[:saltLen], data[saltLen:]
	key, err := h.deriveKey([]byte(secretKey), salt)
	if err != nil {
		return "", err
	}

	block, _ := aes.NewCipher(key)
	gcm, _ := c.NewGCM(block)

	nonceSize := gcm.NonceSize()
	if len(encryptedPayload) < nonceSize {
		return "", errors.New("malformed nonce")
	}

	nonce, ciphertext := encryptedPayload[:nonceSize], encryptedPayload[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

func (h *cryptoHelper) AES256EncryptWithSalt(secretKey, plainText, saltHex string) (string, error) {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return "", errors.New("invalid salt hex")
	}

	key, err := h.deriveKey([]byte(secretKey), salt)
	if err != nil {
		return "", err
	}

	block, _ := aes.NewCipher(key)
	gcm, _ := c.NewGCM(block)

	nonce, err := h.GenerateRandomBytes(gcm.NonceSize())
	if err != nil {
		return "", err
	}

	result := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(result), nil
}

func (h *cryptoHelper) HMACSHA512Sign(secretKey, data string) (string, error) {
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (h *cryptoHelper) HMACSHA512Verify(secretKey, data, expectedHashHex string) bool {
	expectedHash, err := hex.DecodeString(expectedHashHex)
	if err != nil {
		return false
	}

	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write([]byte(data))
	actualHash := mac.Sum(nil)

	return hmac.Equal(actualHash, expectedHash)
}

func (h *cryptoHelper) SHA256Sign(plainText string) (string, error) {
	hash := sha256.Sum256([]byte(plainText))
	return hex.EncodeToString(hash[:]), nil
}

func (h *cryptoHelper) SHA512Sign(plainText string) (string, error) {
	hash := sha512.Sum512([]byte(plainText))
	return hex.EncodeToString(hash[:]), nil
}

func (h *cryptoHelper) GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("invalid length")
	}
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("entropy source failure: %w", err)
	}
	return b, nil
}

func (h *cryptoHelper) GenerateRandomHex(length int) (string, error) {
	b, err := h.GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
