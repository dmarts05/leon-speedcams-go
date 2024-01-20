package speedcamsscraper

import "time"

// Represents the speedcams data for a specific day
type SpeedcamsDayData struct {
	Date      time.Time
	Morning   []Speedcam
	Afternoon []Speedcam
}
