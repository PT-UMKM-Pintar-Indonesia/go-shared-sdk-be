package sdk_dto

type (
	Request[T any] struct {
		Payload T
	}
)
