package speedcamsscraper

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
	dataBuilder := strings.Builder{}

	asterisks := strings.Repeat("*", 31)
	year, month, day := scdd.Date.Date()

	// Using strings.Repeat for repeated characters
	dataBuilder.WriteString(asterisks + "\n")
	dataBuilder.WriteString(fmt.Sprintf("* LEÓN SPEEDCAMS (%02d/%02d/%d) *\n", day, month, year))
	dataBuilder.WriteString(asterisks + "\n\n")

	dataBuilder.WriteString("Morning:\n")
	for _, speedcam := range scdd.Morning {
		dataBuilder.WriteString(fmt.Sprintf("\t- %s\n", speedcam.String()))
	}

	dataBuilder.WriteString("\n")

	dataBuilder.WriteString("Afternoon:\n")
	for _, speedcam := range scdd.Afternoon {
		dataBuilder.WriteString(fmt.Sprintf("\t- %s\n", speedcam.String()))
	}

	return dataBuilder.String()
}
