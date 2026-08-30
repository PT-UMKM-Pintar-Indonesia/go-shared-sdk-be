package sdk_helper

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func NumberToString[T int](n T) string {
	return fmt.Sprintf("%v", n)
}

func ToExponential(f float64, fractionDigits int) string {
	return strconv.FormatFloat(f, 'e', fractionDigits, 64)
}

func ToFixed(f float64, fractionDigits int) string {
	return strconv.FormatFloat(f, 'f', fractionDigits, 64)
}

func ToPrecision(f float64, precision int) string {
	return strconv.FormatFloat(f, 'g', precision, 64)
}

func ValueOf[T int](n T) T {
	return n
}

func IsFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

func IsInteger(f float64) bool {
	return f == math.Trunc(f) && IsFinite(f)
}

func IsNaN(f float64) bool {
	return math.IsNaN(f)
}

func IsSafeInteger(f float64) bool {
	const maxSafeInt = 9007199254740991
	const minSafeInt = -9007199254740991

	return IsInteger(f) && f >= minSafeInt && f <= maxSafeInt
}

func ParseInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func Min[T int](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T int](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Clamp[T int](val, min, max T) T {
	if val < min {
		return min
	}

	if val > max {
		return max
	}

	return val
}
