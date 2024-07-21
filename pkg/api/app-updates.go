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

const IOS = "ios"
const ANDROID = "android"

type UpdateManifest struct {
	ID             string            `json:"id"`
	CreatedAt      string            `json:"createdAt"`
	RuntimeVersion string            `json:"runtimeVersion"`
	LaunchAsset    UpdateAsset       `json:"launchAsset"`
	Assets         []UpdateAsset     `json:"assets"`
	Metadata       map[string]string `json:"metadata"`
	Extra          map[string]any    `json:"extra"`
}

type UpdateAsset struct {
	Hash          string `json:"hash,omitempty"`
	Key           string `json:"key"`
	ContentType   string `json:"contentType"`
	FileExtension string `json:"fileExtension,omitempty"`
	URL           string `json:"url"`
	Path          string `json:"-"`
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
	Extra    map[string]any
	CLI      *config.CLI
}

func (u *Updater) GetManifest(platform, runtime, prefix string) (*UpdateManifest, error) {
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
			Path:          ass.Path,
			URL:           fmt.Sprintf("%s/%s", prefix, ass.Path),
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
		RuntimeVersion: runtime,
		LaunchAsset: UpdateAsset{
			Hash:        hash,
			Key:         dotParts[0],
			Path:        plat.Bundle,
			URL:         fmt.Sprintf("%s/%s", prefix, plat.Bundle),
			ContentType: "application/javascript",
		},
		Assets:   assets,
		Metadata: map[string]string{},
		Extra:    map[string]any{},
	}
	return &man, nil
}

func (u *Updater) GetManifestBytes(platform, runtime, prefix string) ([]byte, error) {
	manifest, err := u.GetManifest(platform, runtime, prefix)
	if err != nil {
		return nil, err
	}
	bs, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

// get MIME types of built-in update files
func (u *Updater) GetMimes() (map[string]string, error) {
	assets := []UpdateAsset{}
	ios, err := u.GetManifest(IOS, "", "")
	if err != nil {
		return nil, err
	}
	assets = append(assets, ios.LaunchAsset)
	assets = append(assets, ios.Assets...)
	android, err := u.GetManifest(ANDROID, "", "")
	if err != nil {
		return nil, err
	}
	assets = append(assets, android.LaunchAsset)
	assets = append(assets, android.Assets...)
	m := map[string]string{}
	for _, ass := range assets {
		if ass.Path == "" {
			return nil, fmt.Errorf("asset has no path! asset=%v", ass)
		}
		m[ass.Path] = ass.ContentType
	}
	return m, nil
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

	file2, err := fs.Open("expoConfig.json")
	if err != nil {
		return nil, err
	}
	bs2, err := io.ReadAll(file2)
	if err != nil {
		return nil, err
	}
	extra := map[string]any{}
	err = json.Unmarshal(bs2, &extra)
	if err != nil {
		return nil, err
	}

	return &Updater{CLI: cli, Metadata: metadata, Extra: extra}, nil
}

func (a *AquareumAPI) HandleAppUpdates(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		prefix := fmt.Sprintf("http://%s", req.Host)
		log.Log(ctx, "got app-updates request", "method", req.Method, "headers", req.Header)
		plat := req.Header.Get("expo-platform")
		if plat == "" {
			log.Log(ctx, "app-updates request missing Expo-Platform")
			w.WriteHeader(400)
			return
		}
		runtime := req.Header.Get("expo-runtime-version")
		if runtime == "" {
			log.Log(ctx, "app-updates request missing Expo-Runtime-Version")
			w.WriteHeader(400)
			return
		}
		bs, err := a.Updater.GetManifestBytes(plat, runtime, prefix)
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

	sEnc := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(outbs)
	return sEnc, nil
}
