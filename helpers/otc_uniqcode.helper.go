package sdk_helper

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"time"
)

var chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateChecksum(input string) string {
	hash := 0
	for _, ch := range input {
		hash = (hash << 5) - hash + int(ch)
		hash |= 0
	}

	index1 := int(math.Abs(float64(hash))) % len(chars)
	index2 := int(math.Abs(float64(hash*13))) % len(chars)

	return string(chars[index1]) + string(chars[index2])
}

func Generate(prefix string) string {
	if prefix != "UMKMS" && prefix != "OTC" {
		prefix = "OTC"
	}

	digits := make([]byte, 4)
	for i := range digits {
		n, err := rand.Int(rand.Reader, big.NewInt(10))

		if err != nil {
			now := time.Now()
			digits = []byte(fmt.Sprintf("%02d%02d", int(now.Month()), now.Day()))
			break
		}

		digits[i] = byte('0' + n.Int64())
	}

	baseCode := prefix + string(digits)

	checksum := generateChecksum(baseCode)

	return baseCode + checksum
}

func Validate(code string) bool {
	matched, _ := regexp.MatchString(`^(UMKMS|OTC)\d{4}[A-Z2-9]{2}$`, code)
	if !matched {
		return false
	}

	baseCode := code[:len(code)-2]
	providedChecksum := code[len(code)-2:]
	calculatedChecksum := generateChecksum(baseCode)

	return providedChecksum == calculatedChecksum
}
