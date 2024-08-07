package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"aquareum.tv/aquareum/pkg/errors"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/mist/misttriggers"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	sloghttp "github.com/samber/slog-http"
)

func (a *AquareumAPI) ServeInternalHTTP(ctx context.Context) error {
	handler, err := a.InternalHandler(ctx)
	if err != nil {
		return err
	}
	return a.ServerWithShutdown(ctx, handler, func(s *http.Server) error {
		s.Addr = a.CLI.HttpInternalAddr
		log.Log(ctx, "http server starting", "addr", s.Addr)
		return s.ListenAndServe()
	})
}

func (a *AquareumAPI) InternalHandler(ctx context.Context) (http.Handler, error) {
	router := httprouter.New()
	broker := misttriggers.NewTriggerBroker()
	broker.OnPushOutStart(func(ctx context.Context, payload *misttriggers.PushOutStartPayload) (string, error) {
		log.Log(ctx, "got push out start", "streamName", payload.StreamName, "url", payload.URL)
		u, err := url.Parse(payload.URL)
		if err != nil {
			return "", err
		}
		uu, err := uuid.NewV7()
		if err != nil {
			return "", fmt.Errorf("error generating uuid: %w", err)
		}
		u.Path, err = url.JoinPath(uu.String(), u.Path)
		if err != nil {
			return "", fmt.Errorf("error joining path: %w", err)
		}

		return u.String(), nil
	})
	triggerCollection := misttriggers.NewMistCallbackHandlersCollection(a.CLI, broker)
	router.POST("/mist-trigger", triggerCollection.Trigger())
	router.POST("/segment/*anything", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Log(ctx, "segment start")
		suffix := strings.TrimPrefix(r.URL.Path, "/")
		err := os.MkdirAll(path.Join(a.CLI.DataDir, path.Dir(suffix)), 0700)
		if err != nil {
			log.Log(ctx, "error making directory", "error", err)
			errors.WriteHTTPInternalServerError(w, "directory create error", err)
			return
		}
		f, err := os.Create(path.Join(a.CLI.DataDir, suffix))
		if err != nil {
			log.Log(ctx, "error opening file", "error", err)
			errors.WriteHTTPInternalServerError(w, "file open error", err)
			return
		}
		defer f.Close()
		fwrite := bufio.NewWriter(f)
		count, err := io.Copy(fwrite, r.Body)
		if err != nil {
			log.Log(ctx, "segment error", "error", err, "len", count)
			errors.WriteHTTPInternalServerError(w, "segment error", err)
			return
		}
		log.Log(ctx, "segment success", "len", count, "url", r.URL.String())
	})
	handler := sloghttp.Recovery(router)
	handler = sloghttp.New(slog.Default())(handler)
	return handler, nil
}
