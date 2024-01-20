package message

import (
	"fmt"
	"net/http"
	"net/url"
)

const telegramApiBaseURL = "https://api.telegram.org"

type TelegramBotMessageSender struct {
	Client *http.Client
	Token  string
	ChatID string
}

func (t TelegramBotMessageSender) SendMessage(message string) error {
	telegramURL := fmt.Sprintf("%s/bot%s/sendMessage?chat_id=%s&text=%s&parse_mode=HTML", telegramApiBaseURL, t.Token, t.ChatID, url.QueryEscape(message))
	if _, err := t.Client.Get(telegramURL); err != nil {
		return fmt.Errorf("error sending message to Telegram: %w", err)
	}

	return nil
}
