package sdk_opt

import (
	"github.com/go-chi/chi/v5"
)

type Graceful struct {
	HANDLER *chi.Mux
}
