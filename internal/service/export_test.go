// export_test.go exposes internal constructors for use by external _test packages.
package service

import (
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// NewModsWithBaseURL creates a Mods service whose HTTP client redirects all
// requests to baseURL. Used by mods_test.go to point at a mock server.
func NewModsWithBaseURL(cfg *config.Config, logger *zap.Logger, baseURL string) *Mods {
	return &Mods{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{
			Timeout:   time.Duration(cfg.Mods.Timeout) * time.Second,
			Transport: &redirectTransport{base: baseURL},
		},
	}
}

// ParseProjectID exposes the unexported parseProjectID for cross-package tests.
func (m *Mods) ParseProjectID(modURL string) (string, error) {
	return m.parseProjectID(modURL)
}

// redirectTransport rewrites the host/scheme of every request to base.
type redirectTransport struct {
	base string
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base, _ := url.Parse(t.base)
	clone := req.Clone(req.Context())
	clone.URL.Scheme = base.Scheme
	clone.URL.Host = base.Host
	clone.Host = base.Host
	return http.DefaultTransport.RoundTrip(clone)
}
