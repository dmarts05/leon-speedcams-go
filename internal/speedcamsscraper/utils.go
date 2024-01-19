package speedcamsscraper

import (
	"fmt"
	"strconv"
)

func filterEmptyStrings(strings []string) []string {
	var filteredStrings []string

	for _, str := range strings {
		if str != "" {
			filteredStrings = append(filteredStrings, str)
		}
	}

	return filteredStrings
}

func stringSliceToIntSlice(stringSlice []string) ([]int, error) {
	var intSlice []int

	for _, str := range stringSlice {
		num, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("failed to convert string slice to int slice: %w", err)
		}

		intSlice = append(intSlice, num)
	}

	return intSlice, nil
}

func filterRowsByDay(rows []SpeedcamsRow, day int) []SpeedcamsRow {
	filteredRows := make([]SpeedcamsRow, 0, 2)
	for _, row := range rows {
		if row.Day == day {
			filteredRows = append(filteredRows, row)
		}
	}
	return filteredRows
}

func appendSpeedcams(existing []Speedcam, streets []string, limits []int) []Speedcam {
	for i, street := range streets {
		existing = append(existing, Speedcam{Street: street, SpeedLimit: limits[i]})
	}
	return existing
}
