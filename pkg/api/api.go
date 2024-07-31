package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	sloghttp "github.com/samber/slog-http"

	"aquareum.tv/aquareum/js/app"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/crypto/signers/eip712"
	apierrors "aquareum.tv/aquareum/pkg/errors"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/model"
)

type AquareumAPI struct {
	CLI     *config.CLI
	Model   model.Model
	Updater *Updater
	Signer  *eip712.EIP712Signer
	Mimes   map[string]string
}

func MakeAquareumAPI(cli *config.CLI, mod model.Model, signer *eip712.EIP712Signer) (*AquareumAPI, error) {
	updater, err := PrepareUpdater(cli)
	if err != nil {
		return nil, err
	}
	a := &AquareumAPI{CLI: cli, Model: mod, Updater: updater, Signer: signer}
	a.Mimes, err = updater.GetMimes()
	if err != nil {
		return nil, err
	}
	return a, nil
}

type AppHostingFS struct {
	http.FileSystem
}

func (fs AppHostingFS) Open(name string) (http.File, error) {
	file, err1 := fs.FileSystem.Open(name)
	if err1 == nil {
		return file, nil
	}
	if !errors.Is(err1, os.ErrNotExist) {
		return nil, err1
	}
	file, err2 := fs.FileSystem.Open(fmt.Sprintf(name + ".html"))
	if err2 == nil {
		return file, nil
	}
	return nil, err1
}

func (a *AquareumAPI) Handler(ctx context.Context) (http.Handler, error) {
	files, err := app.Files()
	if err != nil {
		return nil, err
	}
	router := httprouter.New()
	apiRouter := httprouter.New()
	apiRouter.HandlerFunc("GET", "/api/notification", a.HandleNotification(ctx))
	apiRouter.HandlerFunc("POST", "/api/notification", a.HandleNotification(ctx))
	apiRouter.HandlerFunc("POST", "/api/golive", a.HandleGoLive(ctx))
	// old clients
	router.HandlerFunc("GET", "/app-updates", a.HandleAppUpdates(ctx))
	// new ones
	apiRouter.HandlerFunc("GET", "/api/manifest", a.HandleAppUpdates(ctx))
	apiRouter.NotFound = a.HandleAPI404(ctx)
	router.Handler("GET", "/api/*resource", apiRouter)
	router.Handler("POST", "/api/*resource", apiRouter)
	router.Handler("PUT", "/api/*resource", apiRouter)
	router.Handler("PATCH", "/api/*resource", apiRouter)
	router.Handler("DELETE", "/api/*resource", apiRouter)
	router.Handler("GET", "/dl/*params", a.AppDownloadHandler(ctx))
	router.NotFound = a.FileHandler(ctx, http.FileServer(AppHostingFS{http.FS(files)}))
	handler := sloghttp.Recovery(router)
	handler = cors.Default().Handler(handler)
	handler = sloghttp.New(slog.Default())(handler)

	return handler, nil
}

func (a *AquareumAPI) FileHandler(ctx context.Context, fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		noslash := req.URL.Path[1:]
		ct, ok := a.Mimes[noslash]
		if ok {
			w.Header().Set("content-type", ct)
		}
		fs.ServeHTTP(w, req)
	}
}

func (a *AquareumAPI) RedirectHandler(ctx context.Context) (http.Handler, error) {
	_, tlsPort, err := net.SplitHostPort(a.CLI.HttpsAddr)
	if err != nil {
		return nil, err
	}
	handleRedirect := func(w http.ResponseWriter, req *http.Request) {
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}
		u := req.URL
		if tlsPort == "443" {
			u.Host = host
		} else {
			u.Host = net.JoinHostPort(host, tlsPort)
		}
		u.Scheme = "https"
		http.Redirect(w, req, u.String(), http.StatusTemporaryRedirect)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRedirect)
	return mux, nil
}

type NotificationPayload struct {
	Token string `json:"token"`
}

func (a *AquareumAPI) HandleAPI404(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(404)
	}
}

func (a *AquareumAPI) HandleGoLive(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		payload, err := io.ReadAll(req.Body)
		if err != nil {
			apierrors.WriteHTTPBadRequest(w, "error reading body", err)
			return
		}
		signed, err := a.Signer.Verify(payload)
		if err != nil {
			apierrors.WriteHTTPBadRequest(w, "could not verify signature on payload", err)
		}
		log.Log(ctx, "got signed & verified payload", "payload", signed)
		w.WriteHeader(204)
	}
}

func (a *AquareumAPI) HandleNotification(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			payload, err := io.ReadAll(req.Body)
			if err != nil {
				log.Log(ctx, "error reading notification create", "error", err)
				w.WriteHeader(400)
				return
			}
			n := NotificationPayload{}
			err = json.Unmarshal(payload, &n)
			if err != nil {
				log.Log(ctx, "error unmarshalling notification create", "error", err)
				w.WriteHeader(400)
				return
			}
			err = a.Model.CreateNotification(n.Token)
			if err != nil {
				log.Log(ctx, "error creating notification", "error", err)
				w.WriteHeader(400)
				return
			}
			log.Log(ctx, "successfully created notification", "token", n.Token)
			w.WriteHeader(200)
		} else if req.Method == "GET" {
			// disallow unless we have an admin token
			if a.CLI.AdminSecret == "" {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}
			log.Log(ctx, a.CLI.AdminSecret)
			auth := req.Header.Get("Authorization")
			if auth == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			expected := fmt.Sprintf("Bearer %s", a.CLI.AdminSecret)
			if auth != expected {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			nots, err := a.Model.ListNotifications()
			if err != nil {
				log.Log(ctx, "error listing notifications", "error", err)
				w.WriteHeader(500)
				return
			}
			bs, err := json.Marshal(nots)
			if err != nil {
				log.Log(ctx, "error marshalling notifications", "error", err)
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
			w.Write(bs)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func (a *AquareumAPI) ServeHTTP(ctx context.Context) error {
	handler, err := a.Handler(ctx)
	if err != nil {
		return err
	}
	return a.ServerWithShutdown(ctx, handler, func(s *http.Server) error {
		s.Addr = a.CLI.HttpAddr
		log.Log(ctx, "http server starting", "addr", s.Addr)
		return s.ListenAndServe()
	})
}

func (a *AquareumAPI) ServeHTTPRedirect(ctx context.Context) error {
	handler, err := a.RedirectHandler(ctx)
	if err != nil {
		return err
	}
	return a.ServerWithShutdown(ctx, handler, func(s *http.Server) error {
		s.Addr = a.CLI.HttpAddr
		log.Log(ctx, "http tls redirecct server starting", "addr", s.Addr)
		return s.ListenAndServe()
	})
}

func (a *AquareumAPI) ServeHTTPS(ctx context.Context) error {
	handler, err := a.Handler(ctx)
	if err != nil {
		return err
	}
	return a.ServerWithShutdown(ctx, handler, func(s *http.Server) error {
		s.Addr = a.CLI.HttpsAddr
		log.Log(ctx, "https server starting",
			"addr", s.Addr,
			"certPath", a.CLI.TLSCertPath,
			"keyPath", a.CLI.TLSKeyPath,
		)
		return s.ListenAndServeTLS(a.CLI.TLSCertPath, a.CLI.TLSKeyPath)
	})
}

func (a *AquareumAPI) ServerWithShutdown(ctx context.Context, handler http.Handler, serve func(*http.Server) error) error {
	ctx, cancel := context.WithCancel(ctx)
	handler = gziphandler.GzipHandler(handler)
	server := http.Server{Handler: handler}
	var serveErr error
	go func() {
		serveErr = serve(&server)
		cancel()
	}()
	<-ctx.Done()
	if serveErr != nil {
		return fmt.Errorf("error in http server: %w", serveErr)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}
