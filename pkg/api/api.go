package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"aquareum.tv/aquareum/js/app"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/log"
)

func Handler() (http.Handler, error) {
	mux := http.NewServeMux()
	files, err := app.Files()
	if err != nil {
		return nil, err
	}
	mux.Handle("/", http.FileServer(http.FS(files)))
	return mux, nil
}

func ServeHTTP(ctx context.Context, cli config.CLI) error {
	return ServerWithShutdown(ctx, cli, func(s *http.Server) error {
		s.Addr = cli.HttpAddr
		log.Log(ctx, "http server starting", "addr", s.Addr)
		return s.ListenAndServe()
	})
}

func ServeHTTPS(ctx context.Context, cli config.CLI) error {
	return ServerWithShutdown(ctx, cli, func(s *http.Server) error {
		s.Addr = cli.HttpsAddr
		log.Log(ctx, "https server starting",
			"addr", s.Addr,
			"certPath", cli.TLSCertPath,
			"keyPath", cli.TLSKeyPath,
		)
		return s.ListenAndServeTLS(cli.TLSCertPath, cli.TLSKeyPath)
	})
}

func ServerWithShutdown(ctx context.Context, cli config.CLI, serve func(*http.Server) error) error {
	handler, err := Handler()
	if err != nil {
		return err
	}
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
