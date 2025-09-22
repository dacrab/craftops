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

// Discord color constants
const (
	ColorGreen  = 0x00FF00 // Success
	ColorRed    = 0xFF0000 // Error
	ColorOrange = 0xFFA500 // Warning
	ColorBlue   = 0x0099FF // Info
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
	*BaseService
	client *http.Client
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(cfg *config.Config, logger *zap.Logger) NotificationServiceInterface {
	return &NotificationService{
		BaseService: NewBaseService(cfg, logger),
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

// SendSuccessNotification sends a success notification
func (ns *NotificationService) SendSuccessNotification(ctx context.Context, message string) error {
	if !ns.GetConfig().Notifications.SuccessNotifications {
		return nil
	}
	return ns.sendDiscordNotification(ctx, "✅ Success", message, ColorGreen)
}

// SendErrorNotification sends an error notification
func (ns *NotificationService) SendErrorNotification(ctx context.Context, message string) error {
	if !ns.GetConfig().Notifications.ErrorNotifications {
		return nil
	}
	return ns.sendDiscordNotification(ctx, "❌ Error", message, ColorRed)
}

// SendRestartWarnings sends restart warning notifications
func (ns *NotificationService) SendRestartWarnings(ctx context.Context) error {
	intervals := ns.GetConfig().Notifications.WarningIntervals
	if len(intervals) == 0 {
		return nil
	}

	ns.GetLogger().Info("Sending restart warnings", zap.Ints("intervals", intervals))

	for i, minutes := range intervals {
		warningMsg := strings.ReplaceAll(ns.GetConfig().Notifications.WarningMessage, "{minutes}", fmt.Sprintf("%d", minutes))

		if err := ns.sendDiscordNotification(ctx, "⚠️ Server Restart Warning", warningMsg, ColorOrange); err != nil {
			ns.GetLogger().Error("Failed to send restart warning", zap.Error(err))
			return err
		}

		if i < len(intervals)-1 {
			var waitTime time.Duration
			if i > 0 {
				waitTime = time.Duration(intervals[i-1]-minutes) * time.Minute
			} else {
				waitTime = time.Minute
			}

			ns.GetLogger().Info("Waiting before next warning", zap.Duration("wait_time", waitTime))

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
	if ns.GetConfig().Notifications.DiscordWebhook == "" {
		ns.GetLogger().Debug("Discord webhook not configured, skipping notification")
		return nil
	}

	if ns.GetConfig().DryRun {
		ns.GetLogger().Info("Dry run: Would send Discord notification",
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

	req, err := http.NewRequestWithContext(ctx, "POST", ns.GetConfig().Notifications.DiscordWebhook, bytes.NewBuffer(jsonData))
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

	ns.GetLogger().Debug("Discord notification sent successfully")
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
	if ns.GetConfig().Notifications.DiscordWebhook == "" {
		return HealthCheck{
			Name:    "Discord webhook",
			Status:  "⚠️",
			Message: "Not configured",
		}
	}

	if !strings.HasPrefix(ns.GetConfig().Notifications.DiscordWebhook, "https://discord.com/api/webhooks/") {
		return HealthCheck{
			Name:    "Discord webhook",
			Status:  "❌",
			Message: "Invalid webhook URL format",
		}
	}

	return HealthCheck{
		Name:    "Discord webhook",
		Status:  "✅",
		Message: "Configured",
	}
}

func (ns *NotificationService) checkNotificationSettings() HealthCheck {
	if !ns.GetConfig().Notifications.ErrorNotifications && !ns.GetConfig().Notifications.SuccessNotifications {
		return HealthCheck{
			Name:    "Notification settings",
			Status:  "⚠️",
			Message: "All notifications disabled",
		}
	}

	return HealthCheck{
		Name:    "Notification settings",
		Status:  "✅",
		Message: "Configured",
	}
}
