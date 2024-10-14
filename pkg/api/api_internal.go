package api

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"aquareum.tv/aquareum/pkg/errors"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/media"
	"aquareum.tv/aquareum/pkg/mist/mistconfig"
	"aquareum.tv/aquareum/pkg/mist/misttriggers"
	v0 "aquareum.tv/aquareum/pkg/schema/v0"
	"github.com/julienschmidt/httprouter"
	sloghttp "github.com/samber/slog-http"
	"golang.org/x/sync/errgroup"
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
var mkvRE *regexp.Regexp

func init() {
	mkvRE = regexp.MustCompile(`^\d+\.mkv$`)
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
	router.HandlerFunc("GET", "/healthz", a.HandleHealthz(ctx))

	router.GET("/playback/:user/concat", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		user = strings.ToLower(user)
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
		user = strings.ToLower(user)
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
		user = strings.ToLower(user)
		file := p.ByName("file")
		if file == "" {
			errors.WriteHTTPBadRequest(w, "file required", nil)
			return
		}
		fullpath, err := a.CLI.SegmentFilePath(user, file)
		if err != nil {
			errors.WriteHTTPBadRequest(w, "badly formatted request", err)
			return
		}
		http.ServeFile(w, r, fullpath)
	})

	router.GET("/playback/:user/stream.mkv", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		user = strings.ToLower(user)
		w.Header().Set("Content-Type", "video/x-matroska")
		w.WriteHeader(200)
		err := a.MediaManager.SegmentToMKVPlusOpus(ctx, user, w)
		if err != nil {
			log.Log(ctx, "stream.mkv error", "error", err)
		}
	})

	router.GET("/playback/:user/stream.mp4", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := p.ByName("user")
		if user == "" {
			errors.WriteHTTPBadRequest(w, "user required", nil)
			return
		}
		user = strings.ToLower(user)
		var delayMS int64 = 1000
		userDelay := r.URL.Query().Get("delayms")
		if userDelay != "" {
			var err error
			delayMS, err = strconv.ParseInt(userDelay, 10, 64)
			if err != nil {
				errors.WriteHTTPBadRequest(w, "error parsing delay", err)
				return
			}
			if delayMS > 10000 {
				errors.WriteHTTPBadRequest(w, "delay too large, maximum 10000", nil)
				return
			}
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(200)
		g, ctx := errgroup.WithContext(ctx)
		pr, pw := io.Pipe()
		bufw := bufio.NewWriter(pw)
		g.Go(func() error {
			return a.MediaManager.SegmentToMP4(ctx, user, bufw)
		})
		g.Go(func() error {
			time.Sleep(time.Duration(delayMS) * time.Millisecond)
			_, err := io.Copy(w, pr)
			return err
		})
		g.Wait()
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

	// internal route called for each pushed segment from ffmpeg
	router.POST("/segment/:user/:file", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ms := time.Now().UnixMilli()
		user := p.ByName("user")
		if user == "" {
			log.Log(ctx, "invalid code path: got empty user?")
			errors.WriteHTTPInternalServerError(w, "invalid code path: got empty user?", nil)
			return
		}
		user = strings.ToLower(user)
		f := p.ByName("file")
		if !mkvRE.MatchString(f) {
			errors.WriteHTTPBadRequest(w, "file was not in number.mp4 format", nil)
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

	handleIncomingStream := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Log(ctx, "stream start")
		user, err := a.keyToUser(ctx, p.ByName("key"))
		if err != nil {
			errors.WriteHTTPForbidden(w, "unable to authenticate stream key", err)
			return
		}
		prefix := fmt.Sprintf("%s/segment/%s", a.CLI.OwnInternalURL(), user)
		err = media.SegmentToHTTP(ctx, r.Body, prefix)

		if err != nil {
			log.Log(ctx, "stream error", "error", err)
			errors.WriteHTTPInternalServerError(w, "stream error", err)
			return
		}
		log.Log(ctx, "stream success", "url", r.URL.String())
	}

	// route to accept an incoming mkv stream from OBS, segment it, and push the segments back to this HTTP handler
	router.POST("/stream/:key", handleIncomingStream)
	router.PUT("/stream/:key", handleIncomingStream)
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
	return strings.ToLower(signed.Signer()), nil
}
