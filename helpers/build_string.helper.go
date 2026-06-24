package sdk_helper

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

type bstring struct{}

func NewString() sdk_inf.IBString {
	return &bstring{}
}

func (h *bstring) Capitalize(s string) string {
	if s == sdk_cons.EMPTY {
		return sdk_cons.EMPTY
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])

	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

func (h *bstring) CamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == sdk_cons.EMPTY {
		return sdk_cons.EMPTY
	}

	re := regexp.MustCompile(`[^a-zA-Z0-9]+|\s+`)
	words := re.Split(s, -1)

	var sb strings.Builder
	sb.Grow(len(s))

	for i, word := range words {
		if word == sdk_cons.EMPTY {
			continue
		}
		if i == 0 {
			sb.WriteString(strings.ToLower(word))
		} else {
			sb.WriteString(h.Capitalize(word))
		}
	}

	return sb.String()
}

func (h *bstring) SnakeCase(s string) string {
	s = strings.TrimSpace(s)
	if s == sdk_cons.EMPTY {
		return sdk_cons.EMPTY
	}

	re := regexp.MustCompile(`[^a-zA-Z0-9]+|\s+`)
	s = re.ReplaceAllString(s, " ")

	re = regexp.MustCompile(`([A-Z]+)([A-Z][a-z]|\d)`)
	s = re.ReplaceAllString(s, "$1 $2")
	re = regexp.MustCompile(`([a-z\d])([A-Z])`)
	s = re.ReplaceAllString(s, "$1 $2")

	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")

	return s
}

func (h *bstring) Len(s string) int {
	return utf8.RuneCountInString(s)
}

func (h *bstring) CharAt(s string, index int) string {
	runes := []rune(s)
	if index < 0 || index >= len(runes) {
		return sdk_cons.EMPTY
	}

	return string(runes[index])
}

func (h *bstring) CharCodeAt(s string, index int) (rune, bool) {
	runes := []rune(s)
	if index < 0 || index >= len(runes) {
		return 0, false
	}

	return runes[index], true
}

func (h *bstring) CodePointAt(s string, index int) (rune, bool) {
	return h.CharCodeAt(s, index)
}

func (h *bstring) Concat(strs ...string) string {
	if len(strs) == 0 {
		return sdk_cons.EMPTY
	}
	if len(strs) == 1 {
		return strs[0]
	}

	totalLen := 0
	for _, s := range strs {
		totalLen += len(s)
	}

	var sb strings.Builder
	sb.Grow(totalLen)

	for _, s := range strs {
		sb.WriteString(s)
	}

	return sb.String()
}

func (h *bstring) IndexOf(s string, index int) string {
	runes := []rune(s)
	length := len(runes)

	if index < 0 {
		index += length
	}

	if index < 0 || index >= length {
		return sdk_cons.EMPTY
	}

	return string(runes[index])
}

func (h *bstring) Slice(s string, start, end int) string {
	runes := []rune(s)
	length := len(runes)

	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	if start < 0 {
		start = 0
	}

	if end > length {
		end = length
	}

	if start > end {
		return sdk_cons.EMPTY
	}

	return string(runes[start:end])
}

func (h *bstring) Substr(s string, start, length int) string {
	runes := []rune(s)
	strLen := len(runes)
	if start < 0 {
		start = strLen + start
	}
	if start < 0 {
		start = 0
	}

	if start >= strLen {
		return sdk_cons.EMPTY
	}

	end := start + length
	if length < 0 {
		end = start
	}

	if end > strLen {
		end = strLen
	}

	return string(runes[start:end])
}

func (h *bstring) ToUpperCase(s string) string {
	return strings.ToUpper(s)
}

func (h *bstring) ToLowerCase(s string) string {
	return strings.ToLower(s)
}

func (h *bstring) IsWellFormed(s string) bool {
	return utf8.ValidString(s)
}

func (h *bstring) ToWellFormed(s string) string {
	return strings.ToValidUTF8(s, "\uFFFD")
}

func (h *bstring) TrimStart(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

func (h *bstring) TrimEnd(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func (h *bstring) PadStart(s string, targetLength int, padString string) string {
	if utf8.RuneCountInString(s) >= targetLength {
		return s
	}

	padLen := targetLength - utf8.RuneCountInString(s)
	pad := strings.Repeat(padString, (padLen/utf8.RuneCountInString(padString))+1)

	return pad[:padLen] + s
}

func (h *bstring) PadEnd(s string, targetLength int, padString string) string {
	if utf8.RuneCountInString(s) >= targetLength {
		return s
	}

	padLen := targetLength - utf8.RuneCountInString(s)
	pad := strings.Repeat(padString, (padLen/utf8.RuneCountInString(padString))+1)

	return s + pad[:padLen]
}

func (h *bstring) Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

func (h *bstring) Replace(s, old, new string) string {
	return strings.Replace(s, old, new, 1)
}

func (h *bstring) ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func (h *bstring) Trim(s string) string {
	return strings.TrimSpace(s)
}

func (h *bstring) StartsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func (h *bstring) EndsWith(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func (h *bstring) IndexOfStr(s, substr string) int {
	return strings.Index(s, substr)
}

func (h *bstring) Search(s, pattern string) (int, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return -1, err
	}

	return re.FindStringIndex(s)[0], nil
}

func (h *bstring) Substring(s string, start, end int) string {
	runes := []rune(s)
	length := len(runes)

	if start < 0 {
		start = 0
	}

	if end > length {
		end = length
	}

	if start >= end {
		return sdk_cons.EMPTY
	}

	return string(runes[start:end])
}

func (h *bstring) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func (h *bstring) Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func (h *bstring) Truncate(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	runes := []rune(s)
	return string(runes[:maxLen])
}

func (h *bstring) Compact(s string) string {
	if s == sdk_cons.EMPTY {
		return sdk_cons.EMPTY
	}

	var sb strings.Builder
	sb.Grow(len(s))

	for _, r := range s {
		if !unicode.IsSpace(r) {
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

func (h *bstring) Uniq(s string) string {
	if s == sdk_cons.EMPTY {
		return sdk_cons.EMPTY
	}

	seen := make(map[rune]struct{})

	var sb strings.Builder
	sb.Grow(utf8.RuneCountInString(s))

	for _, r := range s {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

func (h *bstring) Unshift(s string, prefixes ...string) string {
	if len(prefixes) == 0 {
		return s
	}

	totalLen := len(s)
	for _, p := range prefixes {
		totalLen += len(p)
	}

	var sb strings.Builder
	sb.Grow(totalLen)

	for _, p := range prefixes {
		sb.WriteString(p)
	}

	sb.WriteString(s)
	return sb.String()
}

func (h *bstring) Pop(s string) (string, string, bool) {
	if s == "" {
		return "", "", false
	}

	runes := []rune(s)
	if len(runes) == 0 {
		return "", "", false
	}

	lastRune := runes[len(runes)-1]
	remainingString := string(runes[:len(runes)-1])

	return string(lastRune), remainingString, true
}

func (h *bstring) Included(s, substr string) bool {
	return strings.Contains(s, substr)
}

func (h *bstring) Union(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	seen := make(map[rune]struct{})
	var sb strings.Builder

	totalRuneCount := 0
	for _, s := range strs {
		totalRuneCount += utf8.RuneCountInString(s)
	}
	sb.Grow(totalRuneCount)

	for _, s := range strs {
		for _, r := range s {
			if _, ok := seen[r]; !ok {
				seen[r] = struct{}{}
				sb.WriteRune(r)
			}
		}
	}

	return sb.String()
}

func (h *bstring) IncludedIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
