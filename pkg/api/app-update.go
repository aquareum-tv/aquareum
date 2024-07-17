package api

import (
	"context"
	"net/http"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/model"
)

func HandleAppUpdate(ctx context.Context, cli config.CLI, mod model.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Log(ctx, "got app-update request", "method", req.Method, "headers", req.Header)
		w.WriteHeader(501)
	}
}
