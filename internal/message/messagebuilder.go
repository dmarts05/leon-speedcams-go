package message

import (
	"fmt"
	"strings"

	"github.com/dmarts05/leon-speedcams-go/internal/speedcamsscraper"
)

func formatSpeedcam(speedcam speedcamsscraper.Speedcam, monitoredStreets []string) string {
	if speedcam.IsMonitored(monitoredStreets) {
		return fmt.Sprintf("⚠️ <strong> %s: %d km/h </strong> ⚠️", speedcam.Street, speedcam.SpeedLimit)
	}
	return fmt.Sprintf("%s: %d km/h", speedcam.Street, speedcam.SpeedLimit)
}

func BuildSpeedcamsDayDataMessage(speedcamsDayData speedcamsscraper.SpeedcamsDayData, monitoredStreets []string) string {
	sb := strings.Builder{}

	year, month, day := speedcamsDayData.Date.Date()

	asterisks := strings.Repeat("*", 31)
	sb.WriteString(asterisks + "\n")
	sb.WriteString(fmt.Sprintf("* LEÓN SPEEDCAMS (%02d/%02d/%d) *\n", day, month, year))
	sb.WriteString(asterisks + "\n\n")

	sb.WriteString("Morning:\n")
	for _, speedcam := range speedcamsDayData.Morning {
		sb.WriteString(fmt.Sprintf("\t- %s\n", formatSpeedcam(speedcam, monitoredStreets)))
	}

	sb.WriteString("\n")

	sb.WriteString("Afternoon:\n")
	for _, speedcam := range speedcamsDayData.Afternoon {
		sb.WriteString(fmt.Sprintf("\t- %s\n", formatSpeedcam(speedcam, monitoredStreets)))
	}

	return sb.String()
}
