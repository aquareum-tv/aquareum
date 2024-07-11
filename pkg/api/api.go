package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"aquareum.tv/aquareum/js/app"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/model"
)

func Handler(ctx context.Context, mod model.Model) (http.Handler, error) {
	mux := http.NewServeMux()
	files, err := app.Files()
	if err != nil {
		return nil, err
	}
	mux.Handle("/notification", HandleNotificationCreate(ctx, mod))
	mux.Handle("/", http.FileServer(http.FS(files)))
	return mux, nil
}

func RedirectHandler(ctx context.Context, cli config.CLI, mod model.Model) (http.Handler, error) {
	_, tlsPort, err := net.SplitHostPort(cli.HttpsAddr)
	if err != nil {
		return nil, err
	}
	handleRedirect := func(w http.ResponseWriter, req *http.Request) {
		host, _, _ := net.SplitHostPort(req.Host)
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

func HandleNotificationCreate(ctx context.Context, mod model.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
		err = mod.CreateNotification(n.Token)
		if err != nil {
			log.Log(ctx, "error creating notification", "error", err)
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(200)
	}
}

func ServeHTTP(ctx context.Context, cli config.CLI, mod model.Model) error {
	handler, err := Handler(ctx, mod)
	if err != nil {
		return err
	}
	return ServerWithShutdown(ctx, handler, cli, mod, func(s *http.Server) error {
		s.Addr = cli.HttpAddr
		log.Log(ctx, "http server starting", "addr", s.Addr)
		return s.ListenAndServe()
	})
}

func ServeHTTPRedirect(ctx context.Context, cli config.CLI, mod model.Model) error {
	handler, err := RedirectHandler(ctx, cli, mod)
	if err != nil {
		return err
	}
	return ServerWithShutdown(ctx, handler, cli, mod, func(s *http.Server) error {
		s.Addr = cli.HttpAddr
		log.Log(ctx, "http tls redirecct server starting", "addr", s.Addr)
		return s.ListenAndServe()
	})
}

func ServeHTTPS(ctx context.Context, cli config.CLI, mod model.Model) error {
	handler, err := Handler(ctx, mod)
	if err != nil {
		return err
	}
	return ServerWithShutdown(ctx, handler, cli, mod, func(s *http.Server) error {
		s.Addr = cli.HttpsAddr
		log.Log(ctx, "https server starting",
			"addr", s.Addr,
			"certPath", cli.TLSCertPath,
			"keyPath", cli.TLSKeyPath,
		)
		return s.ListenAndServeTLS(cli.TLSCertPath, cli.TLSKeyPath)
	})
}

func ServerWithShutdown(ctx context.Context, handler http.Handler, cli config.CLI, mod model.Model, serve func(*http.Server) error) error {
	ctx, cancel := context.WithCancel(ctx)
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
