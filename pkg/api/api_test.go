package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/model"
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
			cli := config.CLI{HttpAddr: tt.httpAddr, HttpsAddr: tt.httpsAddr}
			mod := model.DBModel{}

			handler, err := RedirectHandler(context.Background(), cli, &mod)
			if err != nil {
				t.Fatalf("RedirectHandler() error = %v", err)
			}

			req := httptest.NewRequest("GET", tt.requestURL, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			result := rr.Result()
			if result.StatusCode != http.StatusTemporaryRedirect {
				t.Errorf("handler returned wrong status code: got %v want %v",
					result.StatusCode, http.StatusTemporaryRedirect)
			}

			redirectURL, err := result.Location()
			if err != nil {
				t.Fatalf("Failed to get redirect location: %v", err)
			}

			if redirectURL.String() != tt.expectedURL {
				t.Errorf("handler returned unexpected redirect URL: got %v want %v",
					redirectURL.String(), tt.expectedURL)
			}
		})
	}
}
