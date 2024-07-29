package misttriggers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/errors"
	"aquareum.tv/aquareum/pkg/log"
	"github.com/julienschmidt/httprouter"
)

const (
	TRIGGER_PUSH_END        = "PUSH_END"
	TRIGGER_PUSH_OUT_START  = "PUSH_OUT_START"
	TRIGGER_PUSH_REWRITE    = "PUSH_REWRITE"
	TRIGGER_STREAM_BUFFER   = "STREAM_BUFFER"
	TRIGGER_LIVE_TRACK_LIST = "LIVE_TRACK_LIST"
	TRIGGER_USER_NEW        = "USER_NEW"
	TRIGGER_USER_END        = "USER_END"
	TRIGGER_STREAM_SOURCE   = "STREAM_SOURCE"
)

type MistCallbackHandlersCollection struct {
	cli    *config.CLI
	broker TriggerBroker
}

func NewMistCallbackHandlersCollection(cli *config.CLI, b TriggerBroker) *MistCallbackHandlersCollection {
	return &MistCallbackHandlersCollection{cli: cli, broker: b}
}

func (d *MistCallbackHandlersCollection) Trigger() httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		payload, err := io.ReadAll(req.Body)
		if err != nil {
			errors.WriteHTTPBadRequest(w, "Cannot read trigger payload", err)
			return
		}

		var requestID = "MistTrigger-" + config.RandomTrailer(8)
		triggerName := req.Header.Get("X-Trigger")
		mistVersion := req.Header.Get("X-Version")
		if mistVersion == "" {
			mistVersion = req.UserAgent()
		}
		ctx := log.WithLogValues(context.Background(),
			"request_id", requestID,
			"trigger_name", triggerName,
			"mist_version", mistVersion,
		)

		body := MistTriggerBody(payload)

		switch triggerName {
		case TRIGGER_PUSH_OUT_START:
			d.TriggerPushOutStart(ctx, w, req, body)
		case TRIGGER_PUSH_END:
			d.TriggerPushEnd(ctx, w, req, body)
		case TRIGGER_STREAM_BUFFER:
			d.TriggerStreamBuffer(ctx, w, req, body)
		case TRIGGER_PUSH_REWRITE:
			d.TriggerPushRewrite(ctx, w, req, body)
		case TRIGGER_LIVE_TRACK_LIST:
			d.TriggerLiveTrackList(ctx, w, req, body)
		case TRIGGER_USER_NEW:
			d.TriggerUserNew(ctx, w, req, body)
		case TRIGGER_USER_END:
			d.TriggerUserEnd(ctx, w, req, body)
		case TRIGGER_STREAM_SOURCE:
			d.TriggerStreamSource(ctx, w, req, body)
		default:
			errors.WriteHTTPBadRequest(w, "Unsupported X-Trigger", fmt.Errorf("unknown trigger '%s'", triggerName))
			return
		}
	}
}

type MistTriggerBody string

func (b MistTriggerBody) Lines() []string {
	trimmed := strings.TrimSpace(string(b))
	lines := strings.Split(trimmed, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return lines
}
