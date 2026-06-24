package sdk_inf

type IBString interface {
	Capitalize(s string) string
	CamelCase(s string) string
	SnakeCase(s string) string
	Len(s string) int
	CharAt(s string, index int) string
	CharCodeAt(s string, index int) (rune, bool)
	CodePointAt(s string, index int) (rune, bool)
	Concat(strs ...string) string
	IndexOf(s string, index int) string
	Slice(s string, start, end int) string
	Substr(s string, start, length int) string
	ToUpperCase(s string) string
	ToLowerCase(s string) string
	IsWellFormed(s string) bool
	ToWellFormed(s string) string
	TrimStart(s string) string
	TrimEnd(s string) string
	PadStart(s string, targetLength int, padString string) string
	PadEnd(s string, targetLength int, padString string) string
	Repeat(s string, count int) string
	Replace(s string, old, new string) string
	ReplaceAll(s string, old, new string) string
	Trim(s string) string
	StartsWith(s, prefix string) bool
	EndsWith(s, suffix string) bool
	IndexOfStr(s, substr string) int
	Search(s, pattern string) (int, error)
	Substring(s string, start, end int) string
	Split(s, sep string) []string
	Unshift(s string, prefixes ...string) string
	Pop(s string) (string, string, bool)
	Included(s, substr string) bool
	Union(strs ...string) string
	Reverse(s string) string
	Truncate(s string, maxLen int) string
	Compact(s string) string
	Uniq(s string) string
	IncludedIgnoreCase(s, substr string) bool
}
