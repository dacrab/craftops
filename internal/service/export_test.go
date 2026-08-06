package service

import (
	"strings"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// NewModsWithBaseURL creates a Mods service that targets baseURL (for tests).
func NewModsWithBaseURL(cfg *config.Config, logger *zap.Logger, baseURL string) *Mods {
	m := NewMods(cfg, logger)
	m.baseURL = strings.TrimRight(baseURL, "/")
	return m
}

// ParseProjectID exposes parseProjectID for cross-package tests.
func ParseProjectID(modURL string) (string, error) {
	return parseProjectID(modURL)
}
