package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// NotificationService handles notification operations
type NotificationService struct {
	config *config.Config
	logger *zap.Logger
	client *http.Client
}

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Color       int                    `json:"color"`
	Timestamp   string                 `json:"timestamp"`
	Footer      map[string]interface{} `json:"footer"`
}

// DiscordWebhookPayload represents a Discord webhook payload
type DiscordWebhookPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

// Discord color constants
const (
	ColorGreen  = 0x00FF00 // Success
	ColorRed    = 0xFF0000 // Error
	ColorOrange = 0xFFA500 // Warning
	ColorBlue   = 0x0099FF // Info
)

// NewNotificationService creates a new notification service instance
func NewNotificationService(cfg *config.Config, logger *zap.Logger) *NotificationService {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &NotificationService{
		config: cfg,
		logger: logger,
		client: client,
	}
}

// HealthCheck performs health checks for the notification service
func (ns *NotificationService) HealthCheck(ctx context.Context) []HealthCheck {
	checks := []HealthCheck{}

	// Check Discord webhook configuration
	if ns.config.Notifications.DiscordWebhook != "" {
		checks = append(checks, HealthCheck{
			Name:    "Discord webhook",
			Status:  "✅",
			Message: "Configured",
		})

		// Test webhook URL format
		if strings.HasPrefix(ns.config.Notifications.DiscordWebhook, "https://discord.com/api/webhooks/") {
			checks = append(checks, HealthCheck{
				Name:    "Discord connectivity",
				Status:  "✅",
				Message: "Webhook URL format valid",
			})
		} else {
			checks = append(checks, HealthCheck{
				Name:    "Discord connectivity",
				Status:  "⚠️",
				Message: "Invalid webhook URL format",
			})
		}
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Discord webhook",
			Status:  "⚠️",
			Message: "Not configured",
		})
	}

	// Check notification settings
	if len(ns.config.Notifications.WarningIntervals) > 0 {
		intervalCount := len(ns.config.Notifications.WarningIntervals)
		checks = append(checks, HealthCheck{
			Name:    "Warning intervals",
			Status:  "✅",
			Message: fmt.Sprintf("%d intervals configured", intervalCount),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Warning intervals",
			Status:  "⚠️",
			Message: "No warning intervals configured",
		})
	}

	return checks
}

// SendSuccessNotification sends a success notification
func (ns *NotificationService) SendSuccessNotification(ctx context.Context, message string) error {
	if !ns.config.Notifications.SuccessNotifications {
		return nil
	}

	return ns.sendDiscordNotification(ctx, "✅ Success", message, ColorGreen)
}

// SendErrorNotification sends an error notification
func (ns *NotificationService) SendErrorNotification(ctx context.Context, message string) error {
	if !ns.config.Notifications.ErrorNotifications {
		return nil
	}

	return ns.sendDiscordNotification(ctx, "❌ Error", message, ColorRed)
}

// SendRestartWarnings sends restart warning notifications
func (ns *NotificationService) SendRestartWarnings(ctx context.Context) error {
	if len(ns.config.Notifications.WarningIntervals) == 0 {
		return nil
	}

	ns.logger.Info("Sending restart warnings",
		zap.Ints("intervals", ns.config.Notifications.WarningIntervals))

	for i, minutes := range ns.config.Notifications.WarningIntervals {
		// Replace {minutes} placeholder in warning message
		warningMsg := strings.ReplaceAll(ns.config.Notifications.WarningMessage, "{minutes}", fmt.Sprintf("%d", minutes))

		// Send Discord notification
		if err := ns.sendDiscordNotification(ctx, "⚠️ Server Restart Warning", warningMsg, ColorOrange); err != nil {
			ns.logger.Error("Failed to send restart warning", zap.Error(err))
			return err
		}

		// Wait between warnings (except for the last one)
		if i < len(ns.config.Notifications.WarningIntervals)-1 {
			var waitTime time.Duration
			if i > 0 {
				waitTime = time.Duration(ns.config.Notifications.WarningIntervals[i-1]-minutes) * time.Minute
			} else {
				waitTime = time.Minute // Default 1 minute wait
			}

			ns.logger.Info("Waiting before next warning", zap.Duration("wait_time", waitTime))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				// Continue to next warning
			}
		}
	}

	return nil
}

// sendDiscordNotification sends a notification to Discord webhook
func (ns *NotificationService) sendDiscordNotification(ctx context.Context, title, message string, color int) error {
	if ns.config.Notifications.DiscordWebhook == "" {
		ns.logger.Debug("Discord webhook not configured, skipping notification")
		return nil
	}

	if ns.config.DryRun {
		ns.logger.Info("Dry run: Would send Discord notification",
			zap.String("title", title),
			zap.String("message", message))
		return nil
	}

	// Truncate message if too long
	if len(message) > 2000 {
		message = message[:1997] + "..."
	}

	// Create Discord embed
	embed := DiscordEmbed{
		Title:       title,
		Description: message,
		Color:       color,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: map[string]interface{}{
			"text": "Minecraft Mod Manager",
		},
	}

	// Create webhook payload
	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{embed},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", ns.config.Notifications.DiscordWebhook, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := ns.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("discord API returned status %d", resp.StatusCode)
	}

	ns.logger.Debug("Discord notification sent successfully")
	return nil
}
