// Module for scraping speedcams data from the given website (https://ahoraleon.com)
package speedcamsscraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dmarts05/leon-speedcams-go/internal/spanishdate"

	log "github.com/sirupsen/logrus"
)

// Filters the given rows by the given day
func filterRowsByDay(rows []SpeedcamsRow, day int) []SpeedcamsRow {
	filteredRows := make([]SpeedcamsRow, 0, 2)
	for _, row := range rows {
		if row.Day == day {
			filteredRows = append(filteredRows, row)
		}
	}
	return filteredRows
}

// Contains the client and the base request URL required to scrape speedcams data
// from the given website
type SpeedcamsScraper struct {
	Client         *http.Client
	BaseRequestURL string
}

// Gets the latest speedcams data link from the website
// Returns an error if the request fails or if the link is not found
func (ss SpeedcamsScraper) getLatestSpeedcamsLink() (string, error) {
	monthName := spanishdate.GetCurrentSpanishMonth()
	requestURL := fmt.Sprintf("%s/?s=radar+%s", ss.BaseRequestURL, monthName)

	res, err := ss.Client.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("failed to get latest speedcams link: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New("failed to get latest speedcams link: request failed with status code: " + strconv.Itoa(res.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to get latest speedcams link: %w", err)
	}

	speedcamsDataLink := doc.Find(".entry-title a").First().AttrOr("href", "")
	if speedcamsDataLink == "" {
		return "", errors.New("failed to get latest speedcams link: no speedcams data link found")
	}

	return speedcamsDataLink, nil
}

// Gets the speedcams data rows from the given body
// Returns an error if the rows are not found
func (ss SpeedcamsScraper) getSpeedcamsRowsFromBody(body io.Reader) ([]SpeedcamsRow, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to get speedcams rows from website body: %w", err)
	}

	rows, err := parseSpeedcamsRows(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to get speedcams rows from website body: %w", err)
	}

	return rows, nil
}

// Gets the speedcams data rows from the given link
// Returns an error if the request fails or if the rows are not found
func (ss SpeedcamsScraper) getSpeedcamsRowsFromLink(link string) ([]SpeedcamsRow, error) {
	res, err := ss.Client.Get(link)
	if err != nil {
		return nil, fmt.Errorf("failed to get speedcams rows from link: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get speedcams rows from link: request failed with status code: %d", res.StatusCode)
	}

	rows, err := ss.getSpeedcamsRowsFromBody(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get speedcams rows from link: %w", err)
	}

	return rows, nil
}

// Gets today's speedcams data from the website
// Returns an error if the request fails or if the data is not found
func (ss SpeedcamsScraper) GetTodaySpeedcamsData() (SpeedcamsDayData, error) {
	speedcamsDataLink, err := ss.getLatestSpeedcamsLink()
	if err != nil {
		return SpeedcamsDayData{}, fmt.Errorf("failed to get today's speedcams data: %w", err)
	}

	log.Debug("Latest speedcams data link: ", speedcamsDataLink)

	speedcamRows, err := ss.getSpeedcamsRowsFromLink(speedcamsDataLink)
	if err != nil {
		return SpeedcamsDayData{}, fmt.Errorf("failed to get today's speedcams data: %w", err)
	}

	// Filter speedcams data rows to get today's data
	today := time.Now().Day()
	todayRows := filterRowsByDay(speedcamRows, today)

	todaysSpeedcamsData := SpeedcamsDayData{Date: time.Now()}

	// Get today's speedcams data
	for _, row := range todayRows {
		switch row.Shift {
		case Morning:
			todaysSpeedcamsData.Morning = append(todaysSpeedcamsData.Morning, row.Speedcams...)
		case Afternoon:
			todaysSpeedcamsData.Afternoon = append(todaysSpeedcamsData.Afternoon, row.Speedcams...)
		default:
			log.Warn("Unexpected shift value: ", row.Shift)
		}
	}

	return todaysSpeedcamsData, nil
}
