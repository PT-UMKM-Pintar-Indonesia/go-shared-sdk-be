package sdk_dto

import (
	"context"
	"time"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

type (
	GenerateClientCredentialsOptions struct {
		Manager      *manage.Manager
		ClientStore  *store.ClientStore
		Ctx          context.Context
		Expired      time.Duration
		ClientID     string
		ClientSecret string
		Scope        string
	}

	GenerateServerClientCredentialsOptions struct {
		Manager      *manage.Manager
		ClientStore  *store.ClientStore
		Expired      time.Duration
		ClientID     string
		ClientSecret string
		Scope        string
	}

	LoadAccessTokenOptions struct {
		Manager      *manage.Manager
		ClientStore  *store.ClientStore
		Server       *server.Server
		Ctx          context.Context
		Token        string
		ClientID     string
		ClientSecret string
		Scope        string
	}
)
