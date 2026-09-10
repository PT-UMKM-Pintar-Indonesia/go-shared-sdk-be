package sdk_inf

type IRandom interface {
	AlphaCharacters(length int) string
	Numeric(length int) string
	Secure(length int, charset string) (string, error)
	Hex(length int) (string, error)
	RandomNumericStr(length int) string
	RandomItemStr(slice []string) string
}
