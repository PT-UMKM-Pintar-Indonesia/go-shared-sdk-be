package sdk_helper

import (
	"fmt"
	"math"
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

	now := time.Now()
	month := fmt.Sprintf("%02d", int(now.Month()))
	day := fmt.Sprintf("%02d", now.Day())
	baseCode := prefix + month + day

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
