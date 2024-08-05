package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"aquareum.tv/aquareum/pkg/log"
)

var (
	re      = regexp.MustCompile(`^aquareum-(v[0-9]+\.[0-9]+\.[0-9]+)(-[0-9a-f]+)?-([0-9a-z]+)-([0-9a-z]+)\.(.+)$`)
	inputRe = regexp.MustCompile(`^aquareum-([0-9a-z]+)-([0-9a-z]+)\.(.+)$`)
)

func (a *AquareumAPI) AppDownloadHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Log(ctx, "got here")
		pathname := r.URL.Path
		parts := strings.Split(pathname, "/")
		if len(parts) < 4 {
			http.Error(w, "usage: /dl/latest/aquareum-linux-arm64.tar.gz", http.StatusBadRequest)
			return
		}

		_, branch, file := parts[1], parts[2], parts[3]
		if branch == "" || file == "" {
			http.Error(w, "usage: /dl/latest/aquareum-linux-arm64.tar.gz", http.StatusBadRequest)
			return
		}

		inputPieces := inputRe.FindStringSubmatch(file)
		if inputPieces == nil {
			http.Error(w, fmt.Sprintf("could not parse filename %s", file), http.StatusBadRequest)
			return
		}

		inputPlatform, inputArch, inputExt := inputPieces[1], inputPieces[2], inputPieces[3]
		packageURL := fmt.Sprintf("%s/packages?order_by=created_at&sort=desc&package_name=%s", a.CLI.GitLabURL, branch)

		packageReq, err := http.Get(packageURL)
		if err != nil {
			http.Error(w, "failed to fetch packages", http.StatusInternalServerError)
			return
		}
		defer packageReq.Body.Close()

		var packages []map[string]interface{}
		if err := json.NewDecoder(packageReq.Body).Decode(&packages); err != nil {
			http.Error(w, "failed to decode package response", http.StatusInternalServerError)
			return
		}

		if len(packages) == 0 {
			http.Error(w, fmt.Sprintf("package for branch %s not found", branch), http.StatusNotFound)
			return
		}

		pkg := packages[0]
		fileURL := fmt.Sprintf("%s/packages/%v/package_files", a.CLI.GitLabURL, pkg["id"])

		fileReq, err := http.Get(fileURL)
		if err != nil {
			http.Error(w, "failed to fetch files", http.StatusInternalServerError)
			return
		}
		defer fileReq.Body.Close()

		var files []map[string]interface{}
		if err := json.NewDecoder(fileReq.Body).Decode(&files); err != nil {
			http.Error(w, "failed to decode file response", http.StatusInternalServerError)
			return
		}

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
			http.Error(w, fmt.Sprintf("could not find a file for platform=%s arch=%s ext=%s", inputPlatform, inputArch, inputExt), http.StatusNotFound)
			return
		}

		http.Redirect(w, r, outURL, http.StatusFound)
	}
}
