package sdk_inf

type IJwtTransform interface {
	Transform(secretKey string, plainText string, rotate int) ([]byte, error)
	Untransform(secretKey string, cipherText string, rotate int) ([]byte, error)
}
