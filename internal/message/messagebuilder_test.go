package message

import (
	"strings"
	"testing"
	"time"

	"github.com/dmarts05/leon-speedcams-go/internal/speedcamsscraper"
)

func TestFormatSpeedcam(t *testing.T) {
	speedcam := speedcamsscraper.Speedcam{Street: "Test Street", SpeedLimit: 50}
	monitoredStreets := []string{"Test Street"}

	result := formatSpeedcam(speedcam, monitoredStreets)

	expected := "⚠️ <strong> Test Street: 50 km/h </strong> ⚠️"
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	speedcam = speedcamsscraper.Speedcam{Street: "Another Street", SpeedLimit: 50}
	monitoredStreets = []string{"Test Street"}

	result = formatSpeedcam(speedcam, monitoredStreets)

	expected = "Another Street: 50 km/h"
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestBuildSpeedcamsDayDataMessage(t *testing.T) {
	speedcamsDayData := speedcamsscraper.SpeedcamsDayData{
		Date:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Morning:   []speedcamsscraper.Speedcam{{Street: "Test Street", SpeedLimit: 50}},
		Afternoon: []speedcamsscraper.Speedcam{{Street: "Another Street", SpeedLimit: 50}},
	}
	monitoredStreets := []string{"Test Street"}

	result := BuildSpeedcamsDayDataMessage(speedcamsDayData, monitoredStreets)

	expected := strings.Repeat("*", 31) + "\n" +
		"* LEÓN SPEEDCAMS (01/01/2023) *\n" +
		strings.Repeat("*", 31) + "\n\n" +
		"Morning:\n" +
		"\t- ⚠️ <strong> Test Street: 50 km/h </strong> ⚠️\n\n" +
		"Afternoon:\n" +
		"\t- Another Street: 50 km/h\n"

	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}
