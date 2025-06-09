package sdk_dto

type (
	Request[T any] struct {
		Action string
		Req    T
		Body   T
		Param  T
		Query  T
		Option T
		Config T
	}
)
