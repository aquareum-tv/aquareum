package api

import (
	"bytes"
	"context"
	"encoding/base64"
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
var segmentRE *regexp.Regexp

func init() {
	segmentRE = regexp.MustCompile(`^\/segment\/([a-z0-9-\.]+)_([0-9]+)\/([0-9]+)\.ts$`)
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
	router.POST("/mist-segment/*anything", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
		buf := bytes.Buffer{}
		err = media.MuxToMP4(ctx, r.Body, &buf)
		reader := bytes.NewReader(buf.Bytes())
		media.SignMP4(ctx, reader, f)
		if err != nil {
			log.Log(ctx, "segment error", "error", err)
			errors.WriteHTTPInternalServerError(w, "segment error", err)
			return
		}
		log.Log(ctx, "segment success", "url", r.URL.String(), "file", segmentFile, "drift", drift)
	})

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

		userDir := path.Join(a.CLI.DataDir, "segments", user)
		err := os.MkdirAll(userDir, 0700)
		if err != nil {
			log.Log(ctx, "error making directory", "error", err)
			errors.WriteHTTPInternalServerError(w, "directory create error", err)
			return
		}
		segmentFile := path.Join(userDir, fmt.Sprintf("%d.mp4", ms))
		buf := bytes.Buffer{}
		err = media.MuxToMP4(ctx, r.Body, &buf)
		if err != nil {
			log.Log(ctx, "mp4 muxing error", "error", err)
			errors.WriteHTTPInternalServerError(w, "segment error", err)
			return
		}
		log.Log(ctx, "got back from MuxToMP4", "len", len(buf.Bytes()))
		reader := bytes.NewReader(buf.Bytes())
		f, err := os.Create(segmentFile)
		if err != nil {
			log.Log(ctx, "error opening file", "error", err)
			errors.WriteHTTPInternalServerError(w, "file open error", err)
			return
		}
		defer f.Close()
		err = media.SignMP4(ctx, reader, f)
		if err != nil {
			log.Log(ctx, "segment error", "error", err)
			errors.WriteHTTPInternalServerError(w, "segment error", err)
			return
		}
		log.Log(ctx, "segment success", "url", r.URL.String(), "file", segmentFile)
	})

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
