package sdk_pkg

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	goauth2 "github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

type oauth2 struct{}

func NewOauth2() sdk_inf.IOauth2 {
	return oauth2{}
}

func (p oauth2) Manager() *manage.Manager {
	return manage.NewDefaultManager()
}

func (p oauth2) Client() *store.ClientStore {
	return store.NewClientStore()
}

func (p oauth2) Server(manager *manage.Manager) *server.Server {
	srv := server.NewDefaultServer(manager)

	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	srv.SetAuthorizeScopeHandler(func(w http.ResponseWriter, r *http.Request) (scope string, err error) {
		return scope, nil
	})

	return srv
}

func (p oauth2) GenerateClientCredentials(req sdk_dto.GenerateClientCredentialsOptions) (goauth2.TokenInfo, error) {
	manager := req.Manager
	client := req.ClientStore

	client.Set(req.ClientID, &models.Client{
		ID:     req.ClientID,
		Secret: req.ClientSecret,
	})

	manager.SetClientTokenCfg(&manage.Config{
		AccessTokenExp:    req.Expired,
		IsGenerateRefresh: false,
	})
	manager.MapClientStorage(client)
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	return manager.GenerateAccessToken(req.Ctx, goauth2.ClientCredentials, &goauth2.TokenGenerateRequest{
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		Scope:        req.Scope,
	})
}

func (p oauth2) GenerateServerClientCredentials(req sdk_dto.GenerateServerClientCredentialsOptions) (goauth2.TokenInfo, *server.Server, error) {
	manager := req.Manager
	client := req.ClientStore

	client.Set(req.ClientID, &models.Client{
		ID:     req.ClientID,
		Secret: req.ClientSecret,
	})

	manager.SetClientTokenCfg(&manage.Config{
		AccessTokenExp:    req.Expired,
		IsGenerateRefresh: false,
	})
	manager.MapClientStorage(client)
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	srv := p.Server(manager)

	srv.SetAllowedGrantType(goauth2.ClientCredentials)
	srv.SetAccessTokenExpHandler(func(w http.ResponseWriter, r *http.Request) (exp time.Duration, err error) {
		return req.Expired, nil
	})

	form := url.Values{}
	form.Set("grant_type", string(goauth2.ClientCredentials))
	form.Set("client_id", req.ClientID)
	form.Set("client_secret", req.ClientSecret)
	form.Set("scope", req.Scope)

	r, err := http.NewRequest(http.MethodPost, "", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, err
	}

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	gt, tgr, err := srv.ValidationTokenRequest(r)
	if err != nil {
		return nil, nil, err
	}

	tokenInfo, err := srv.GetAccessToken(r.Context(), gt, tgr)
	if err != nil {
		return nil, nil, err
	}

	return tokenInfo, srv, nil
}

func (p oauth2) LoadAccessToken(req sdk_dto.LoadAccessTokenOptions) (goauth2.TokenInfo, error) {
	server := req.Server
	return server.Manager.LoadAccessToken(req.Ctx, req.Token)
}
