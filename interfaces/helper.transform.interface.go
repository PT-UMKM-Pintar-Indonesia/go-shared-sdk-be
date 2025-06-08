package sdk_inf

import "context"

type ITransform interface {
	SrcToDest(src, dest any) error
	CtxToStruct(ctx context.Context, key string, dest any) error
	EnvToStruct(name, path, ext string, dest any) error
}
