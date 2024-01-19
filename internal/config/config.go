package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	// RequestTimeout is the timeout in seconds for HTTP requests
	RequestTimeout int `json:"request_timeout"`
	// BaseRequestURL is the base URL for HTTP requests
	BaseRequestURL string `json:"base_request_url"`
	// MonitoredStreets is the list of streets to monitor
	MonitoredStreets []string `json:"monitored_streets"`
	// TelegramBotToken is the token for the Telegram bot
	TelegramBotToken string `json:"telegram_bot_token"`
	// TelegramChatID is the chat ID for the Telegram bot
	TelegramChatID string `json:"telegram_chat_id"`
}

// NewConfig returns a new Config instance by reading the configuration file
func NewConfig() (Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := Config{}
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file: %w", err)
	}

	return config, nil
}
