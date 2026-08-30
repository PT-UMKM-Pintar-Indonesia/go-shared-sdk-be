package sdk_inf

type IArgon2 interface {
	Hash(password []byte) ([]byte, error)
	Verify(password, hashPassword []byte) (bool, error)
}
