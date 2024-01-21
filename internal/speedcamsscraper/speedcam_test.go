package speedcamsscraper

import "testing"

func TestSpeedcam_IsMonitored(t *testing.T) {
	speedcam := Speedcam{Street: "Test Street", SpeedLimit: 50}
	monitoredStreets := []string{"Test Street"}

	result := speedcam.IsMonitored(monitoredStreets)

	if !result {
		t.Errorf("Expected '%t', but got '%t'", true, result)
	}

	speedcam = Speedcam{Street: "Another Street", SpeedLimit: 50}
	monitoredStreets = []string{"Test Street"}

	result = speedcam.IsMonitored(monitoredStreets)

	if result {
		t.Errorf("Expected '%t', but got '%t'", false, result)
	}
}
