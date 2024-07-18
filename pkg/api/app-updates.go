package api

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"aquareum.tv/aquareum/js/app"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/log"
)

const RUNTIME_VERSION = "0.0.2"
const IOS = "ios"
const ANDROID = "android"

type UpdateManifest struct {
	ID             string            `json:"id"`
	CreatedAt      string            `json:"createdAt"`
	RuntimeVersion string            `json:"runtimeVersion"`
	LaunchAsset    UpdateAsset       `json:"launchAsset"`
	Assets         []UpdateAsset     `json:"assets"`
	Metadata       map[string]string `json:"metadata"`
	Extra          map[string]string `json:"extra"`
}

type UpdateAsset struct {
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

type Updater struct {
	Metadata ExpoMetadata
	CLI      *config.CLI
}

func (u *Updater) GetManifest(platform string) (*UpdateManifest, error) {
	var plat ExpoMetadataPlatform
	if platform == IOS {
		plat = u.Metadata.FileMetadata.IOS
	} else if platform == ANDROID {
		plat = u.Metadata.FileMetadata.Android
	} else {
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}
	assets := []UpdateAsset{}
	for _, ass := range plat.Assets {
		ext := fmt.Sprintf(".%s", ass.Ext)
		typ := mime.TypeByExtension(ext)
		if typ == "" {
			return nil, fmt.Errorf("unknown content-type for file extention %s", ext)
		}
		parts := strings.Split(ass.Path, "/")
		hash, err := hashFile(ass.Path)
		if err != nil {
			return nil, err
		}
		assets = append(assets, UpdateAsset{
			Hash:          hash,
			Key:           parts[len(parts)-1],
			URL:           fmt.Sprintf("https://980b-24-19-207-220.ngrok-free.app/%s", ass.Path),
			ContentType:   typ,
			FileExtension: ass.Ext,
		})
	}
	dashParts := strings.Split(plat.Bundle, "-")
	dotParts := strings.Split(dashParts[len(dashParts)-1], ".")
	hash, err := hashFile(plat.Bundle)
	if err != nil {
		return nil, err
	}
	man := UpdateManifest{
		ID:             u.CLI.Build.UUID,
		CreatedAt:      u.CLI.Build.BuildTimeStrExpo(),
		RuntimeVersion: RUNTIME_VERSION,
		LaunchAsset: UpdateAsset{
			Hash:        hash,
			Key:         dotParts[0],
			URL:         fmt.Sprintf("https://980b-24-19-207-220.ngrok-free.app/%s", plat.Bundle),
			ContentType: "application/javascript",
		},
		Assets:   assets,
		Metadata: map[string]string{},
		Extra:    map[string]string{},
	}
	return &man, nil
}

func (u *Updater) GetManifestBytes(platform string) ([]byte, error) {
	manifest, err := u.GetManifest(platform)
	if err != nil {
		return nil, err
	}
	bs, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func PrepareUpdater(cli *config.CLI) (*Updater, error) {
	fs, err := app.Files()
	if err != nil {
		return nil, err
	}
	file, err := fs.Open("metadata.json")
	if err != nil {
		return nil, err
	}
	bs, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	metadata := ExpoMetadata{}
	err = json.Unmarshal(bs, &metadata)
	if err != nil {
		return nil, err
	}
	return &Updater{CLI: cli, Metadata: metadata}, nil
}

func (a *AquareumAPI) HandleAppUpdates(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Log(ctx, "got app-updates request", "method", req.Method, "headers", req.Header)
		plat := req.Header.Get("expo-platform")
		if plat == "" {
			log.Log(ctx, "app-updates request missing Expo-Platform")
			w.WriteHeader(400)
			return
		}
		bs, err := a.Updater.GetManifestBytes(plat)
		if err != nil {
			log.Log(ctx, "app-updates request errored getting manfiest", "error", err)
			w.WriteHeader(400)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Header().Set("expo-protocol-version", "1")
		w.Header().Set("expo-sfv-version", "0")
		w.WriteHeader(http.StatusOK)
		w.Write(bs)
	}
}

func hashFile(path string) (string, error) {
	fs, err := app.Files()
	if err != nil {
		return "", err
	}
	file, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	bs, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	h := sha256.New()

	h.Write(bs)

	outbs := h.Sum(nil)

	sEnc := base64.StdEncoding.EncodeToString(outbs)
	return sEnc, nil
}
