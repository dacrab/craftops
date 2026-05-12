package service

import (
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// NewModsWithBaseURL creates a Mods service that redirects requests to baseURL (for tests).
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

// ParseProjectID exposes parseProjectID for cross-package tests.
func ParseProjectID(modURL string) (string, error) {
	return parseProjectID(modURL)
}

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
