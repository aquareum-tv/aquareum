package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"aquareum.tv/aquareum/js/app"
	"aquareum.tv/aquareum/pkg/log"
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

type ExpoMetadata struct {
	Version      int    `json:"version"`
	Bundler      string `json:"bundler"`
	FileMetadata struct {
		IOS     ExpoMetadataPlatform `json:"ios"`
		Android ExpoMetadataPlatform `json:"android"`
	} `json:"fileMetadata"`
}

type ExpoMetadataPlatform struct {
	Bundle string `json:"bundle"`
	Assets []struct {
		Path string `json:"path"`
		Ext  string `json:"ext"`
	} `json:"assets"`
}

// func init() {
// 	err := InitUpdater()
// 	if err != nil {
// 		panic(err)
// 	}
// }

func InitUpdater() error {
	fs, err := app.Files()
	if err != nil {
		panic(err)
	}
	file, err := fs.Open("metadata.json")
	if err != nil {
		return err
	}
	bs, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	metadata := ExpoMetadata{}
	err = json.Unmarshal(bs, &metadata)
	if err != nil {
		return err
	}
	panic(fmt.Sprintf("%v", metadata))
	return nil
}

func (a *AquareumAPI) HandleAppUpdates(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Log(ctx, "got app-updates request", "method", req.Method, "headers", req.Header)
		w.WriteHeader(501)
	}
}
