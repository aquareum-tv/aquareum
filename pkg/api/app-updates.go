package api

import (
	"context"
	"net/http"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/model"
)

type Manifest struct {
	ID             string            `json:"id"`
	CreatedAt      string            `json:"createdAt"`
	RuntimeVersion string            `json:"runtimeVersion"`
	LaunchAsset    Asset             `json:"launchAsset"`
	Assets         []Asset           `json:"assets"`
	Metadata       map[string]string `json:"metadata"`
	Extra          map[string]string `json:"extra"`
}

type Asset struct {
	Hash          string `json:"hash,omitempty"`
	Key           string `json:"key"`
	ContentType   string `json:"contentType"`
	FileExtension string `json:"fileExtension,omitempty"`
	URL           string `json:"url"`
}

func HandleAppUpdates(ctx context.Context, cli config.CLI, mod model.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Log(ctx, "got app-updates request", "method", req.Method, "headers", req.Header)
		w.WriteHeader(501)
	}
}
