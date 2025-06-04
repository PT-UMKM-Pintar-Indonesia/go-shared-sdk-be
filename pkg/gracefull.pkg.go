package pkg

import (
	"crypto/tls"
	"net/http"
	"os"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
	"github.com/ory/graceful"
)

func Graceful(req sdk_dto.Request[sdk_dto.Environtment], Handler func() sdk_opt.Graceful) error {
	h := Handler()
	secure := true

	if _, ok := os.LookupEnv("GO_ENV"); ok && req.Config.APP.ENV != sdk_cons.DEV {
		secure = false
	}

	server := http.Server{
		Handler:        h.HANDLER,
		Addr:           ":" + req.Config.APP.PORT,
		MaxHeaderBytes: req.Config.APP.INBOUND_SIZE,
		TLSConfig:      &tls.Config{InsecureSkipVerify: secure},
	}

	Logrus(sdk_cons.INFO, "Server listening on port %s", req.Config.APP.PORT)
	return graceful.Graceful(server.ListenAndServe, server.Shutdown)
}
