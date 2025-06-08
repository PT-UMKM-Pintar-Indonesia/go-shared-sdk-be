package sdk_inf

import (
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

type IOauth2 interface {
	Manager() *manage.Manager
	Client() *store.ClientStore
	Server(*manage.Manager) *server.Server
	GenerateClientCredentials(req sdk_dto.GenerateClientCredentialsOptions) (oauth2.TokenInfo, error)
	GenerateServerClientCredentials(req sdk_dto.GenerateServerClientCredentialsOptions) (oauth2.TokenInfo, *server.Server, error)
	LoadAccessToken(req sdk_dto.LoadAccessTokenOptions) (oauth2.TokenInfo, error)
}
