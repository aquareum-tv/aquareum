package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"aquareum.tv/aquareum/pkg/errors"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/media"
	"aquareum.tv/aquareum/pkg/mist/mistconfig"
	"aquareum.tv/aquareum/pkg/mist/misttriggers"
	v0 "aquareum.tv/aquareum/pkg/schema/v0"
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

// lightweight way to authenticate push requests to ourself
var secretUUID string

func init() {
	uu, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	secretUUID = uu.String()
}

func (a *AquareumAPI) InternalHandler(ctx context.Context) (http.Handler, error) {
	router := httprouter.New()
	broker := misttriggers.NewTriggerBroker()
	broker.OnPushOutStart(func(ctx context.Context, payload *misttriggers.PushOutStartPayload) (string, error) {
		return payload.URL, nil
	})
	broker.OnPushRewrite(func(ctx context.Context, payload *misttriggers.PushRewritePayload) (string, error) {
		log.Log(ctx, "got push out start", "streamName", payload.StreamName, "url", payload.URL.String())

		ms := time.Now().UnixMilli()
		out := fmt.Sprintf("%s+%s_%d", mistconfig.STREAM_NAME, payload.StreamName, ms)

		return out, nil
	})
	triggerCollection := misttriggers.NewMistCallbackHandlersCollection(a.CLI, broker)
	router.POST("/mist-trigger", triggerCollection.Trigger())

	router.GET("/playback/:user/concat", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		w.Header().Set("content-type", "text/plain")
		fmt.Fprintf(w, "ffconcat version 1.0\n")
		// intermittent reports that you need two here to make things work properly? shouldn't matter.
		for i := 0; i < 2; i += 1 {
			fmt.Fprintf(w, "file '%s/playback/%s/latest.mp4'\n", a.CLI.OwnInternalURL(), user)
		}
	})

	router.GET("/playback/:user/latest.mp4", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		log.Log(ctx, "got latest.mp4 request", "user", user)
		file := <-a.MediaManager.SubscribeSegment(ctx, user)
		w.Header().Set("Location", fmt.Sprintf("%s/playback/%s/segment/%s\n", a.CLI.OwnInternalURL(), user, file))
		w.WriteHeader(301)
	})

	router.GET("/playback/:user/segment/:file", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		file := p.ByName("file")
		if file == "" {
			errors.WriteHTTPBadRequest(w, "file required", nil)
			return
		}
		fullpath := filepath.Join(a.CLI.DataDir, "segments", user, file)
		http.ServeFile(w, r, fullpath)
	})

	router.GET("/playback/:user/stream.mkv", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		w.Header().Set("Content-Type", "video/x-matroska")
		w.WriteHeader(200)
		err := a.MediaManager.StreamToMKV(ctx, user, w)
		if err != nil {
			log.Log(ctx, "stream.mkv error", "error", err)
		}
	})

	router.HEAD("/playback/:user/stream.mkv", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		w.Header().Set("Content-Type", "video/x-matroska")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(200)
	})

	// handler for post-segmented mkv streams
	router.POST("/playback/:user/:uuid/stream.mkv", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		uu := p.ByName("uuid")
		if uu == "" {
			errors.WriteHTTPBadRequest(w, "uuid required", nil)
			return
		}
		a.MediaManager.HandleMKVStream(ctx, user, uu, r.Body)
	})

	// internal route called for each pushed segment from ffmpeg
	router.POST("/segment/:uuid/:user/:file", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ms := time.Now().UnixMilli()
		uu := p.ByName("uuid")
		if uu != secretUUID {
			errors.WriteHTTPForbidden(w, "unable to authenticate internal url", nil)
			return
		}
		user := p.ByName("user")
		if user == "" {
			log.Log(ctx, "invalid code path: got empty user?")
			errors.WriteHTTPInternalServerError(w, "invalid code path: got empty user?", nil)
			return
		}
		ctx := log.WithLogValues(ctx, "user", user, "file", p.ByName("file"), "time", fmt.Sprintf("%d", ms))
		err := a.MediaManager.SignSegment(ctx, r.Body, ms)
		if err != nil {
			log.Log(ctx, "segment error", "error", err)
			errors.WriteHTTPInternalServerError(w, "segment error", err)
			return
		}
	})

	// route to accept an incoming mkv stream from OBS, segment it, and push the segments back to this HTTP handler
	router.POST("/stream/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Log(ctx, "stream start")
		user, err := a.keyToUser(ctx, p.ByName("key"))
		if err != nil {
			errors.WriteHTTPForbidden(w, "unable to authenticate stream key", err)
			return
		}
		prefix := fmt.Sprintf("%s/segment/%s/%s", a.CLI.OwnInternalURL(), secretUUID, user)
		err = media.SegmentToHTTP(ctx, r.Body, prefix)

		if err != nil {
			log.Log(ctx, "stream error", "error", err)
			errors.WriteHTTPInternalServerError(w, "stream error", err)
			return
		}
		log.Log(ctx, "stream success", "url", r.URL.String())
	})
	handler := sloghttp.Recovery(router)
	handler = sloghttp.New(slog.Default())(handler)
	return handler, nil
}

func (a *AquareumAPI) keyToUser(ctx context.Context, key string) (string, error) {
	payload, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	signed, err := a.Signer.Verify(payload)
	if err != nil {
		return "", err
	}
	_, ok := signed.Data().(*v0.StreamKey)
	if !ok {
		return "", fmt.Errorf("got signed data but it wasn't a stream key")
	}
	return signed.Signer(), nil
}
