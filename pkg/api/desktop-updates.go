package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"aquareum.tv/aquareum/pkg/aqtime"
	apierrors "aquareum.tv/aquareum/pkg/errors"
	"aquareum.tv/aquareum/pkg/log"
	"github.com/julienschmidt/httprouter"
)

func formatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}

type MacManifestUpdateTo struct {
	Version string `json:"version"`
	PubDate string `json:"pub_date"`
	Notes   string `json:"notes"`
	Name    string `json:"name"`
	URL     string `json:"url"`
}

type MacManifestRelease struct {
	Version  string              `json:"version"`
	UpdateTo MacManifestUpdateTo `json:"updateTo"`
}

type MacManifest struct {
	CurrentRelease string               `json:"currentRelease"`
	Releases       []MacManifestRelease `json:"releases"`
}

func (a *AquareumAPI) HandleDesktopUpdates(ctx context.Context) httprouter.Handle {
	mac := a.HandleMacDesktopUpdates(ctx)
	win := a.HandleWindowsDesktopUpdates(ctx)
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		platform := params.ByName("platform")
		if platform == "darwin" {
			mac(w, req, params)
		} else if platform == "windows" {
			win(w, req, params)
		} else {
			apierrors.WriteHTTPBadRequest(w, fmt.Sprintf("unsupported platform: %s", platform), nil)
		}
	}
}

func (a *AquareumAPI) HandleMacDesktopUpdates(ctx context.Context) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		platform := params.ByName("platform")
		architecture := params.ByName("architecture")
		clientVersion := params.ByName("version")
		clientBuildTime := params.ByName("buildTime")
		file := params.ByName("file")
		if file != "RELEASES.json" {
			apierrors.WriteHTTPNotFound(w, fmt.Sprintf("unknown file: %s", file), nil)
			return
		}
		log.Log(ctx, formatRequest(req),
			"platform", platform,
			"architecture", architecture,
			"clientVersion", clientVersion,
			"clientBuildTime", clientBuildTime,
		)
		clientBuildSec, err := strconv.ParseInt(clientBuildTime, 10, 64)
		if err != nil {
			apierrors.WriteHTTPBadRequest(w, "build time must be a number", err)
			return
		}
		var mani MacManifest
		if clientBuildSec >= a.CLI.Build.BuildTime {
			// client is newer or the same as server
			mani = MacManifest{
				CurrentRelease: clientVersion,
				Releases:       []MacManifestRelease{},
			}
		} else {
			// we're newer than the client, tell it to update
			aqt := aqtime.FromSec(a.CLI.Build.BuildTime)
			// sigh. but at least it's only for dev versions.
			serverVersionZ := strings.ReplaceAll(a.CLI.Build.Version, "-", "-z")
			updateTo := MacManifestUpdateTo{
				Version: serverVersionZ,
				PubDate: aqt.String(),
				Notes:   fmt.Sprintf("Aquareum %s", clientVersion),
				Name:    fmt.Sprintf("Aquareum %s", clientVersion),
				URL:     fmt.Sprintf("https://%s/dl/latest/aquareum-desktop-%s-%s.zip", req.Host, platform, architecture),
			}

			mani = MacManifest{
				CurrentRelease: serverVersionZ,
				Releases: []MacManifestRelease{
					{
						Version:  clientVersion,
						UpdateTo: updateTo,
					},
					// todo: this is straight from their example, but why does this version upgrade to itself...?
					{
						Version:  serverVersionZ,
						UpdateTo: updateTo,
					},
				},
			}
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		bs, err := json.Marshal(mani)
		if err != nil {
			log.Log(ctx, "error marshaling mac update manifest", "error", err)
		}
		w.Write(bs)
	}
}

func (a *AquareumAPI) HandleWindowsDesktopUpdates(ctx context.Context) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		platform := params.ByName("platform")
		architecture := params.ByName("architecture")
		clientVersion := params.ByName("version")
		clientBuildTime := params.ByName("buildTime")
		file := params.ByName("file")
		log.Log(ctx, formatRequest(req),
			"platform", platform,
			"architecture", architecture,
			"clientVersion", clientVersion,
			"clientBuildTime", clientBuildTime,
		)
		clientBuildSec, err := strconv.ParseInt(clientBuildTime, 10, 64)
		if err != nil {
			apierrors.WriteHTTPBadRequest(w, "build time must be a number", err)
			return
		}

		if file == "RELEASES" {
			if clientBuildSec >= a.CLI.Build.BuildTime {
				// client is newer or the same as server
				fmt.Fprintf(w, "0000000000000000000000000000000000000000 aquareum_desktop-%s-full.nupkg 1", clientVersion)
				return
			}
			fmt.Fprintf(w, "1CBC2208DECB3E55C7AEA7320258AA36E3297F18 aquareum_desktop-0.1.4-full.nupkg 174710960")
			return
		}
		http.Redirect(w, req, "https://git.aquareum.tv/api/v4/projects/1/packages/generic/electron/v0.1.3-5742a5a4/aquareum-desktop-v0.1.3-5742a5a4-windows-amd64.1cbc2208decb3e55c7aea7320258aa36e3297f18.nupkg", http.StatusTemporaryRedirect)
	}
}
