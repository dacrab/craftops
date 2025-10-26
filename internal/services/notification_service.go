package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// Discord color constants
const (
	ColorGreen  = 0x00FF00 // Success
	ColorRed    = 0xFF0000 // Error
	ColorOrange = 0xFFA500 // Warning
)

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

// NotificationService handles notification operations
type NotificationService struct {
	config *config.Config
	logger *zap.Logger
	client *http.Client
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(cfg *config.Config, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		config: cfg,
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendSuccessNotification sends a success notification
func (ns *NotificationService) SendSuccessNotification(ctx context.Context, message string) error {
	if !ns.config.Notifications.SuccessNotifications {
		return nil
	}
    return ns.sendDiscordNotification(ctx, "Success", message, ColorGreen)
}

// SendErrorNotification sends an error notification
func (ns *NotificationService) SendErrorNotification(ctx context.Context, message string) error {
	if !ns.config.Notifications.ErrorNotifications {
		return nil
	}
    return ns.sendDiscordNotification(ctx, "Error", message, ColorRed)
}

// SendRestartWarnings sends restart warning notifications
func (ns *NotificationService) SendRestartWarnings(ctx context.Context) error {
	intervals := append([]int(nil), ns.config.Notifications.WarningIntervals...)
	if len(intervals) == 0 {
		return nil
	}
	// Ensure descending order (e.g., 15,10,5,1)
	sort.Slice(intervals, func(i, j int) bool { return intervals[i] > intervals[j] })

	ns.logger.Info("Sending restart warnings", zap.Ints("intervals", intervals))

	for i, minutes := range intervals {
		warningMsg := strings.ReplaceAll(ns.config.Notifications.WarningMessage, "{minutes}", fmt.Sprintf("%d", minutes))

        if err := ns.sendDiscordNotification(ctx, "Server Restart Warning", warningMsg, ColorOrange); err != nil {
			ns.logger.Error("Failed to send restart warning", zap.Error(err))
			return err
		}

		if i < len(intervals)-1 {
			next := intervals[i+1]
			diff := minutes - next
			if diff < 0 {
				diff = 0
			}
			waitTime := time.Duration(diff) * time.Minute

			ns.logger.Info("Waiting before next warning", zap.Duration("wait_time", waitTime))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
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

	if len(message) > 2000 {
		message = message[:1997] + "..."
	}

	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{{
			Title:       title,
			Description: message,
			Color:       color,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Footer:      map[string]interface{}{"text": "CraftOps"},
		}},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ns.config.Notifications.DiscordWebhook, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ns.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("discord API returned status %d", resp.StatusCode)
	}

	ns.logger.Debug("Discord notification sent successfully")
	return nil
}

// HealthCheck performs health checks for the notification service
func (ns *NotificationService) HealthCheck(ctx context.Context) []HealthCheck {
	checks := make([]HealthCheck, 0, 2)

	// Check Discord webhook configuration
	checks = append(checks, ns.checkDiscordWebhook())
	// Check notification settings
	checks = append(checks, ns.checkNotificationSettings())

	return checks
}

func (ns *NotificationService) checkDiscordWebhook() HealthCheck {
	if ns.config.Notifications.DiscordWebhook == "" {
		return HealthCheck{
			Name:    "Discord webhook",
            Status:  "WARN",
			Message: "Not configured",
		}
	}

	if !strings.HasPrefix(ns.config.Notifications.DiscordWebhook, "https://discord.com/api/webhooks/") {
		return HealthCheck{
			Name:    "Discord webhook",
            Status:  "ERROR",
			Message: "Invalid webhook URL format",
		}
	}

	return HealthCheck{
		Name:    "Discord webhook",
        Status:  "OK",
		Message: "Configured",
	}
}

func (ns *NotificationService) checkNotificationSettings() HealthCheck {
	if !ns.config.Notifications.ErrorNotifications && !ns.config.Notifications.SuccessNotifications {
		return HealthCheck{
			Name:    "Notification settings",
            Status:  "WARN",
			Message: "All notifications disabled",
		}
	}

	return HealthCheck{
		Name:    "Notification settings",
        Status:  "OK",
		Message: "Configured",
	}
}
