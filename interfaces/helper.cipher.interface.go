package sdk_inf

type ICipher interface {
	Base64Encode(plainText string) string
	Base64Decode(cipherText string) string
	EncodeRotation(plainText string) string
	DecodeRotation(cipherText string) string
	CaesarEncrypt(plainText string, rotation int) string
	CaesarDecrypt(cipherText string, rotation int) string
	RotateNumber(plainNumber string, shift int) string
	DerotateNumber(cipherNumber string, shift int) string
}
