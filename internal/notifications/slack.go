package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color      string       `json:"color,omitempty"`
	Title      string       `json:"title,omitempty"`
	Text       string       `json:"text,omitempty"`
	Fields     []SlackField `json:"fields,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	Timestamp  int64        `json:"ts,omitempty"`
	MarkdownIn []string     `json:"mrkdwn_in,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SendSlackNotification sends a notification to Slack via webhook
func SendSlackNotification(webhookURL, title, message string, count, threshold int) error {
	// Determine color based on severity
	color := getAlertColor(count, threshold)

	// Create rich Slack message
	slackMsg := SlackMessage{
		Username:  "Peep",
		IconEmoji: ":rotating_light:",
		Attachments: []SlackAttachment{
			{
				Color: color,
				Title: fmt.Sprintf("🚨 Alert: %s", title),
				Text:  message,
				Fields: []SlackField{
					{
						Title: "Count",
						Value: fmt.Sprintf("%d", count),
						Short: true,
					},
					{
						Title: "Threshold",
						Value: fmt.Sprintf("%d", threshold),
						Short: true,
					},
					{
						Title: "Severity",
						Value: getSeverityText(count, threshold),
						Short: true,
					},
				},
				Footer:     "Peep Observability",
				Timestamp:  time.Now().Unix(),
				MarkdownIn: []string{"text", "fields"},
			},
		},
	}

	return sendSlackWebhook(webhookURL, slackMsg)
}

// SendSlackMessage sends a simple text message to Slack
func SendSlackMessage(webhookURL, message string) error {
	slackMsg := SlackMessage{
		Text:      message,
		Username:  "Peep",
		IconEmoji: ":mag:",
	}

	return sendSlackWebhook(webhookURL, slackMsg)
}

// sendSlackWebhook sends the actual HTTP request to Slack
func sendSlackWebhook(webhookURL string, message SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send Slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// getAlertColor returns appropriate color based on alert severity
func getAlertColor(count, threshold int) string {
	ratio := float64(count) / float64(threshold)

	switch {
	case ratio >= 3.0:
		return "danger" // Red
	case ratio >= 2.0:
		return "warning" // Orange
	case ratio >= 1.5:
		return "#ffcc00" // Yellow
	default:
		return "good" // Green
	}
}

// getSeverityText returns human-readable severity
func getSeverityText(count, threshold int) string {
	ratio := float64(count) / float64(threshold)

	switch {
	case ratio >= 3.0:
		return "🔴 Critical"
	case ratio >= 2.0:
		return "🟠 High"
	case ratio >= 1.5:
		return "🟡 Medium"
	default:
		return "🟢 Low"
	}
}
