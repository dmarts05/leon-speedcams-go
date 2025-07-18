package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// EnableCron indicates whether the cron job is enabled
	EnableCron bool
	// Cron is the cron expression for scheduling tasks
	Cron string
	// RequestTimeout is the timeout in seconds for HTTP requests
	RequestTimeout int
	// BaseRequestURL is the base URL for HTTP requests
	BaseRequestURL string
	// MonitoredStreets is the list of streets to monitor
	MonitoredStreets []string
	// TelegramBotToken is the token for the Telegram bot
	TelegramBotToken string
	// TelegramChatID is the chat ID for the Telegram bot
	TelegramChatID string
}

// New returns a new Config instance by reading environment variables
func New() (Config, error) {
	enableCronStr := os.Getenv("ENABLE_CRON")
	enableCron := enableCronStr == "1" || strings.ToLower(enableCronStr) == "true"

	requestTimeoutStr := os.Getenv("REQUEST_TIMEOUT")
	if requestTimeoutStr == "" {
		return Config{}, fmt.Errorf("missing required environment variable: REQUEST_TIMEOUT")
	}
	requestTimeout, err := strconv.Atoi(requestTimeoutStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid REQUEST_TIMEOUT value: %w", err)
	}

	monitoredStreets := strings.Split(os.Getenv("MONITORED_STREETS"), ",")

	config := Config{
		EnableCron:       enableCron,
		Cron:             os.Getenv("CRON"),
		RequestTimeout:   requestTimeout,
		BaseRequestURL:   os.Getenv("BASE_REQUEST_URL"),
		MonitoredStreets: monitoredStreets,
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}

	// Validate required fields
	missing := []string{}
	if config.EnableCron && config.Cron == "" {
		missing = append(missing, "CRON")
	}
	if config.BaseRequestURL == "" {
		missing = append(missing, "BASE_REQUEST_URL")
	}
	if len(config.MonitoredStreets) == 0 || config.MonitoredStreets[0] == "" {
		missing = append(missing, "MONITORED_STREETS")
	}
	if config.TelegramBotToken == "" {
		missing = append(missing, "TELEGRAM_BOT_TOKEN")
	}
	if config.TelegramChatID == "" {
		missing = append(missing, "TELEGRAM_CHAT_ID")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return config, nil
}
