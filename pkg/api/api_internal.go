package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"aquareum.tv/aquareum/pkg/errors"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/media"
	"aquareum.tv/aquareum/pkg/mist/mistconfig"
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

var segmentRE *regexp.Regexp

func init() {
	segmentRE = regexp.MustCompile(`^\/segment\/([a-z0-9-\.]+)_([0-9]+)\/([0-9]+)\.ts$`)
}

func (a *AquareumAPI) InternalHandler(ctx context.Context) (http.Handler, error) {
	router := httprouter.New()
	broker := misttriggers.NewTriggerBroker()
	broker.OnPushOutStart(func(ctx context.Context, payload *misttriggers.PushOutStartPayload) (string, error) {
		// aquareum://$wildcard/$currentMediaTime.ts?split=1&video=maxbps&audio=AAC&append=1
		// log.Log(ctx, "got push out start", "streamName", payload.StreamName, "url", payload.URL)
		// u, err := url.Parse(payload.URL)
		// if err != nil {
		// 	return "", err
		// }
		// u.Path, err = url.JoinPath(uu.String(), u.Path)
		// if err != nil {
		// 	return "", fmt.Errorf("error joining path: %w", err)
		// }

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
	router.POST("/segment/*anything", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Log(ctx, "segment start")
		ms := time.Now().UnixMilli()
		matches := segmentRE.FindStringSubmatch(r.URL.Path)
		if len(matches) != 4 {
			log.Log(ctx, "regex failed on /segment/url", "path", r.URL.Path)
			errors.WriteHTTPInternalServerError(w, "segment error", nil)
			return
		}
		user := matches[1]
		startTimeStr := matches[2]
		mediaTimeStr := matches[3]
		startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			log.Log(ctx, "error parsing number", "error", err)
			errors.WriteHTTPInternalServerError(w, "error parsing number", err)
			return
		}
		mediaTime, err := strconv.ParseInt(mediaTimeStr, 10, 64)
		if err != nil {
			log.Log(ctx, "error parsing number", "error", err)
			errors.WriteHTTPInternalServerError(w, "error parsing number", err)
			return
		}
		segmentTime := startTime + mediaTime
		drift := ms - segmentTime

		if err != nil {
			log.Log(ctx, "error parsing segment start time", "error", err, "startTime", startTimeStr)
			errors.WriteHTTPInternalServerError(w, "segment error", err)
			return
		}
		userDir := path.Join(a.CLI.DataDir, "segments", user)
		err = os.MkdirAll(userDir, 0700)
		if err != nil {
			log.Log(ctx, "error making directory", "error", err)
			errors.WriteHTTPInternalServerError(w, "directory create error", err)
			return
		}
		segmentFile := path.Join(userDir, fmt.Sprintf("%d.mp4", segmentTime))
		f, err := os.Create(segmentFile)
		if err != nil {
			log.Log(ctx, "error opening file", "error", err)
			errors.WriteHTTPInternalServerError(w, "file open error", err)
			return
		}
		defer f.Close()
		err = media.MuxToMP4(ctx, r.Body, f)
		if err != nil {
			log.Log(ctx, "segment error", "error", err)
			errors.WriteHTTPInternalServerError(w, "segment error", err)
			return
		}
		log.Log(ctx, "segment success", "url", r.URL.String(), "file", segmentFile, "drift", drift)
	})
	handler := sloghttp.Recovery(router)
	handler = sloghttp.New(slog.Default())(handler)
	return handler, nil
}
