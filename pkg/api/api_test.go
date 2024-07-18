package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name        string
		httpAddr    string
		httpsAddr   string
		requestURL  string
		expectedURL string
	}{
		{
			name:        "default https port",
			httpAddr:    "0.0.0.0:80",
			httpsAddr:   "0.0.0.0:443",
			requestURL:  "http://example.com/",
			expectedURL: "https://example.com/",
		},
		{
			name:        "non-default https port",
			httpAddr:    "0.0.0.0:80",
			httpsAddr:   "0.0.0.0:8443",
			requestURL:  "http://example.com/",
			expectedURL: "https://example.com:8443/",
		},
		{
			name:        "non-default http port",
			httpAddr:    "0.0.0.0:8080",
			httpsAddr:   "0.0.0.0:443",
			requestURL:  "http://example.com:8080/",
			expectedURL: "https://example.com/",
		},
		{
			name:        "non-default both",
			httpAddr:    "0.0.0.0:8080",
			httpsAddr:   "0.0.0.0:8443",
			requestURL:  "http://example.com:8080/",
			expectedURL: "https://example.com:8443/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &config.CLI{HttpAddr: tt.httpAddr, HttpsAddr: tt.httpsAddr}
			mod := &model.DBModel{}
			a := AquareumAPI{CLI: cli, Mod: mod}

			handler, err := a.RedirectHandler(context.Background())
			assert.NoError(t, err, "RedirectHandler should not return an error")

			req := httptest.NewRequest("GET", tt.requestURL, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			result := rr.Result()
			assert.Equal(t, http.StatusTemporaryRedirect, result.StatusCode, "handler returned wrong status code")

			redirectURL, err := result.Location()
			assert.NoError(t, err, "Failed to get redirect location")

			assert.Equal(t, tt.expectedURL, redirectURL.String(), "handler returned unexpected redirect URL")
		})
	}
}
