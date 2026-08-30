package sdk_helper

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
)

var weakPins = map[string]bool{
	"123456": true, "123456789": true, "12345678": true, "12345": true,
	"000000": true, "111111": true, "222222": true, "333333": true,
	"444444": true, "555555": true, "666666": true, "777777": true,
	"888888": true, "999999": true, "121212": true, "123123": true,
}

func SecurePin(pin string) (bool, error) {
	if err := ValidatePin(pin); err != nil {
		return false, err
	}

	if weakPins[pin] {
		return false, nil
	}

	if hasSequentialOrRepeated(pin) {
		return false, nil
	}

	if isDatePattern(pin) {
		return false, nil
	}

	return true, nil
}

func ValidatePin(pin string) error {
	return ValidatePinWithOptions(pin, 4, 6, false)
}

func ValidatePinWithOptions(pin string, min, max int, requireAlpha bool) error {
	pin = strings.TrimSpace(pin)
	n := len(pin)

	if n < min || n > max {
		return fmt.Errorf("PIN length must be between %d and %d", min, max)
	}

	hasDigit := false
	hasSpecial := false

	for i := 0; i < n; i++ {
		c := pin[i]
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}

	if !hasDigit {
		return errors.New("PIN must contain digits")
	}

	if requireAlpha && !hasSpecial {
		return errors.New("PIN must contain at least one non-digit character")
	}

	return nil
}

func hasSequentialOrRepeated(pin string) bool {
	n := len(pin)
	if n < 3 {
		return false
	}

	ascCount, descCount, repeatCount := 1, 1, 1

	for i := 1; i < n; i++ {
		diff := int(pin[i]) - int(pin[i-1])

		if diff == 1 {
			ascCount++
			descCount = 1
			repeatCount = 1
		} else if diff == -1 {
			descCount++
			ascCount = 1
			repeatCount = 1
		} else if diff == 0 {
			repeatCount++
			ascCount = 1
			descCount = 1
		} else {
			ascCount, descCount, repeatCount = 1, 1, 1
		}

		if ascCount >= 3 || descCount >= 3 || repeatCount >= 3 {
			return true
		}
	}
	return false
}

func isDatePattern(pin string) bool {
	n := len(pin)
	currYear := time.Now().Year()

	for i := 0; i <= n-4; i++ {
		year := 0
		for j := 0; j < 4; j++ {
			year = year*10 + int(pin[i+j]-'0')
		}
		if year >= 1950 && year <= currYear+1 {
			return true
		}
	}

	if n >= 4 {
		p1 := int(pin[0]-'0')*10 + int(pin[1]-'0')
		p2 := int(pin[2]-'0')*10 + int(pin[3]-'0')

		isMonthDay := (p1 >= 1 && p1 <= 12) && (p2 >= 1 && p2 <= 31)
		isDayMonth := (p2 >= 1 && p2 <= 12) && (p1 >= 1 && p1 <= 31)

		if isMonthDay || isDayMonth {
			return true
		}
	}

	return false
}

func GenerateSecurePIN(length int) (string, error) {
	if length < 4 || length > 12 {
		return "", errors.New("invalid length")
	}

	const digits = "0123456789"
	res := make([]byte, length)
	randomBytes := make([]byte, length)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		res[i] = digits[randomBytes[i]%10]
	}

	return string(res), nil
}

func CleanPin(raw string) string {
	var sb strings.Builder
	sb.Grow(len(raw))

	for i := 0; i < len(raw); i++ {
		if raw[i] >= '0' && raw[i] <= '9' {
			sb.WriteByte(raw[i])
		}
	}

	return sb.String()
}
