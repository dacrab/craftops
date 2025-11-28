package service

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
	"craftops/internal/domain"
)

const (
	colorGreen  = 0x00FF00
	colorRed    = 0xFF0000
	colorOrange = 0xFFA500
)

// Notification implements Notifier
type Notification struct {
	cfg    *config.Config
	logger *zap.Logger
	client *http.Client
}

var _ Notifier = (*Notification)(nil)

// NewNotification creates a new notification service
func NewNotification(cfg *config.Config, logger *zap.Logger) *Notification {
	return &Notification{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendSuccess sends a success notification
func (n *Notification) SendSuccess(ctx context.Context, message string) error {
	if !n.cfg.Notifications.SuccessNotifications {
		return nil
	}
	return n.sendDiscord(ctx, "Success", message, colorGreen)
}

// SendError sends an error notification
func (n *Notification) SendError(ctx context.Context, message string) error {
	if !n.cfg.Notifications.ErrorNotifications {
		return nil
	}
	return n.sendDiscord(ctx, "Error", message, colorRed)
}

// SendRestartWarnings sends restart warning notifications
func (n *Notification) SendRestartWarnings(ctx context.Context) error {
	intervals := append([]int(nil), n.cfg.Notifications.WarningIntervals...)
	if len(intervals) == 0 {
		return nil
	}

	sort.Slice(intervals, func(i, j int) bool { return intervals[i] > intervals[j] })

	n.logger.Info("Sending restart warnings", zap.Ints("intervals", intervals))

	for i, minutes := range intervals {
		msg := strings.ReplaceAll(n.cfg.Notifications.WarningMessage, "{minutes}", fmt.Sprintf("%d", minutes))

		if err := n.sendDiscord(ctx, "Server Restart Warning", msg, colorOrange); err != nil {
			return err
		}

		if i < len(intervals)-1 {
			next := intervals[i+1]
			wait := time.Duration(minutes-next) * time.Minute

			n.logger.Info("Waiting before next warning", zap.Duration("wait", wait))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
	}

	return nil
}

// HealthCheck performs health checks
func (n *Notification) HealthCheck(_ context.Context) []domain.HealthCheck {
	return []domain.HealthCheck{
		n.checkWebhook(),
		n.checkSettings(),
	}
}

type discordEmbed struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Color       int               `json:"color"`
	Timestamp   string            `json:"timestamp"`
	Footer      map[string]string `json:"footer"`
}

type discordPayload struct {
	Embeds []discordEmbed `json:"embeds"`
}

func (n *Notification) sendDiscord(ctx context.Context, title, message string, color int) error {
	if n.cfg.Notifications.DiscordWebhook == "" {
		n.logger.Debug("Discord webhook not configured, skipping")
		return nil
	}

	if n.cfg.DryRun {
		n.logger.Info("Dry run: Would send Discord notification", zap.String("title", title))
		return nil
	}

	if len(message) > 2000 {
		message = message[:1997] + "..."
	}

	payload := discordPayload{
		Embeds: []discordEmbed{{
			Title:       title,
			Description: message,
			Color:       color,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Footer:      map[string]string{"text": "CraftOps"},
		}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.cfg.Notifications.DiscordWebhook, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return &domain.APIError{
			URL:        n.cfg.Notifications.DiscordWebhook,
			StatusCode: resp.StatusCode,
			Message:    "Discord API error",
		}
	}

	n.logger.Debug("Discord notification sent")
	return nil
}

func (n *Notification) checkWebhook() domain.HealthCheck {
	webhook := n.cfg.Notifications.DiscordWebhook
	if webhook == "" {
		return domain.HealthCheck{Name: "Discord webhook", Status: domain.StatusWarn, Message: "Not configured"}
	}

	if !strings.HasPrefix(webhook, "https://discord.com/api/webhooks/") {
		return domain.HealthCheck{Name: "Discord webhook", Status: domain.StatusError, Message: "Invalid URL format"}
	}

	return domain.HealthCheck{Name: "Discord webhook", Status: domain.StatusOK, Message: "Configured"}
}

func (n *Notification) checkSettings() domain.HealthCheck {
	if !n.cfg.Notifications.ErrorNotifications && !n.cfg.Notifications.SuccessNotifications {
		return domain.HealthCheck{Name: "Notification settings", Status: domain.StatusWarn, Message: "All disabled"}
	}
	return domain.HealthCheck{Name: "Notification settings", Status: domain.StatusOK, Message: "Configured"}
}
