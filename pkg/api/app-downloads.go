package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	apierrors "aquareum.tv/aquareum/pkg/errors"
	"aquareum.tv/aquareum/pkg/log"
	"github.com/julienschmidt/httprouter"
)

var (
	re      = regexp.MustCompile(`^aquareum-(v[0-9]+\.[0-9]+\.[0-9]+)(-[0-9a-f]+)?-([0-9a-z]+)-([0-9a-z]+)\.(.+)$`)
	inputRe = regexp.MustCompile(`^aquareum-([0-9a-z]+)-([0-9a-z]+)\.(.+)$`)
)

func queryGitlabReal(url string) (io.ReadCloser, error) {
	req, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return req.Body, nil
}

var queryGitlab = queryGitlabReal

func (a *AquareumAPI) HandleAppDownload(ctx context.Context) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		log.Log(ctx, "got here")
		pathname := r.URL.Path
		parts := strings.Split(pathname, "/")
		if len(parts) < 4 {
			apierrors.WriteHTTPBadRequest(w, "usage: /dl/latest/aquareum-linux-arm64.tar.gz", nil)
			return
		}

		_, branch, file := parts[1], parts[2], parts[3]
		if branch == "" || file == "" {
			apierrors.WriteHTTPBadRequest(w, "usage: /dl/latest/aquareum-linux-arm64.tar.gz", nil)
			return
		}

		inputPieces := inputRe.FindStringSubmatch(file)
		if inputPieces == nil {
			apierrors.WriteHTTPBadRequest(w, fmt.Sprintf("could not parse filename %s", file), nil)
			return
		}

		inputPlatform, inputArch, inputExt := inputPieces[1], inputPieces[2], inputPieces[3]
		packageURL := fmt.Sprintf("%s/packages?order_by=created_at&sort=desc&package_name=%s", a.CLI.GitLabURL, branch)

		packageBody, err := queryGitlab(packageURL)
		if err != nil {
			apierrors.WriteHTTPInternalServerError(w, "failed to fetch packages", err)
			return
		}
		defer packageBody.Close()

		var packages []map[string]interface{}
		if err := json.NewDecoder(packageBody).Decode(&packages); err != nil {
			apierrors.WriteHTTPInternalServerError(w, "failed to decode package response", err)
			return
		}
		// bs, _ := json.Marshal(packages)
		// fmt.Println(string(bs))

		if len(packages) == 0 {
			apierrors.WriteHTTPNotFound(w, fmt.Sprintf("package for branch %s not found", branch), nil)
			return
		}

		pkg := packages[0]
		fileURL := fmt.Sprintf("%s/packages/%v/package_files", a.CLI.GitLabURL, pkg["id"])

		fileBody, err := queryGitlab(fileURL)
		if err != nil {
			apierrors.WriteHTTPInternalServerError(w, "failed to fetch files", err)
			return
		}
		defer fileBody.Close()

		var files []map[string]interface{}
		if err := json.NewDecoder(fileBody).Decode(&files); err != nil {
			apierrors.WriteHTTPInternalServerError(w, "failed to decode file response", err)
			return
		}
		// bs, _ = json.Marshal(files)
		// fmt.Println(string(bs))

		var foundFile map[string]interface{}
		var outURL string
		for _, f := range files {
			filename := f["file_name"].(string)
			pieces := re.FindStringSubmatch(filename)
			if pieces == nil {
				log.Log(ctx, "could not parse filename %s", "filename", filename)
				continue
			}
			ver, hash, platform, arch, ext := pieces[1], pieces[2], pieces[3], pieces[4], pieces[5]
			if platform == inputPlatform && arch == inputArch && ext == inputExt {
				foundFile = f
				fullVer := ver + hash
				outURL = fmt.Sprintf("%s/packages/generic/%s/%s/%s", a.CLI.GitLabURL, branch, fullVer, filename)
				break
			}
		}

		if foundFile == nil {
			apierrors.WriteHTTPNotFound(w, "could not find a file for platform=%s arch=%s ext=%s", nil)
			return
		}

		http.Redirect(w, r, outURL, http.StatusTemporaryRedirect)
	}
}
