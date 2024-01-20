package speedcamsscraper

import "strings"

// Represents a speedcam in a specific street
type Speedcam struct {
	Street     string
	SpeedLimit int
}

func (s Speedcam) IsMonitored(monitoredStreets []string) bool {
	for _, monitoredStreet := range monitoredStreets {
		if strings.Contains(s.Street, monitoredStreet) {
			return true
		}
	}
	return false
}
