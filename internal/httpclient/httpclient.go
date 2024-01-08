package httpclient

import (
	"net/http"
	"time"
)

// Create an HTTP client with a timeout
func NewHTTPClient(timeout int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
}
