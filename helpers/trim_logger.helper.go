package sdk_helper

import "strings"

const (
	DefaultMaxLogLength = 2048
	TruncatedSuffix     = "...[truncated]"
)

func TrimLogString(s string) string {
	return TrimLogStringWithMax(s, DefaultMaxLogLength)
}

func TrimLogStringWithMax(s string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = DefaultMaxLogLength
	}

	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	suffixLen := len([]rune(TruncatedSuffix))
	cutAt := maxLen - suffixLen
	if cutAt < 0 {
		cutAt = 0
	}

	return string(runes[:cutAt]) + TruncatedSuffix
}

func TrimLogBytes(b []byte, maxLen int) string {
	return TrimLogStringWithMax(string(b), maxLen)
}

func SanitizeLogString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")

	fields := strings.Fields(s)
	s = strings.Join(fields, " ")

	return TrimLogStringWithMax(s, maxLen)
}
