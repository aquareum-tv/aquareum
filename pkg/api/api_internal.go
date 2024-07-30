package api

import (
	"context"
	"log/slog"
	"net/http"

	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/mist/misttriggers"
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
	triggerCollection := misttriggers.NewMistCallbackHandlersCollection(a.CLI, broker)
	router.POST("/mist-trigger", triggerCollection.Trigger())
	handler := sloghttp.Recovery(router)
	handler = sloghttp.New(slog.Default())(handler)
	return handler, nil
}
