// Utility to create an HTTP client with a timeout
package timeoutclient

import (
	"net/http"
	"time"
)

// Create an HTTP client with a timeout
func NewTimeoutClient(timeout int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
}
