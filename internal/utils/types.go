package utils

import (
	"fmt"
	"strings"
	"time"
)

// Enum representing the shift of the day in which the speedcams are active
type Shift int

const (
	Morning Shift = iota
	Afternoon
)

// Row of the speedcams table in the website
type SpeedcamsRow struct {
	Day         int
	Shift       Shift
	Streets     []string
	SpeedLimits []int
}

// Represents a speedcam in a specific street
type Speedcam struct {
	Street     string
	SpeedLimit int
}

// Returns a string in the format "street: speed_limit km/h"
func (sc *Speedcam) String() string {
	return fmt.Sprintf("%s: %d km/h", sc.Street, sc.SpeedLimit)
}

// Represents the speedcams data for a specific day
type SpeedcamsDayData struct {
	Date      time.Time
	Morning   []Speedcam
	Afternoon []Speedcam
}

// Returns a string with all the speedcams data for a specific day
func (scdd *SpeedcamsDayData) String() string {
	sb := strings.Builder{}

	year, month, day := scdd.Date.Date()

	sb.WriteString("*******************************\n")
	sb.WriteString(fmt.Sprintf("* LEÓN SPEEDCAMS (%02d/%02d/%d) *\n", day, month, year))
	sb.WriteString("*******************************\n\n")

	sb.WriteString("Morning:\n")
	for _, speedcam := range scdd.Morning {
		sb.WriteString(fmt.Sprintf("\t- %s\n", speedcam.String()))
	}

	sb.WriteString("\n")

	sb.WriteString("Afternoon:\n")
	for _, speedcam := range scdd.Afternoon {
		sb.WriteString(fmt.Sprintf("\t- %s\n", speedcam.String()))
	}

	return sb.String()
}
