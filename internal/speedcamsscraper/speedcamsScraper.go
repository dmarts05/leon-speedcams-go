// Module for scraping speedcams data from the given website (https://ahoraleon.com)
package speedcamsscraper

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dmarts05/leon-speedcams-go/internal/date"

	log "github.com/sirupsen/logrus"
)

// Contains the client and the base request URL required to scrape speedcams data
// from the given website
type SpeedcamsScraper struct {
	client         *http.Client
	baseRequestURL string
}

// Creates a new SpeedcamsScraper
// The base URL must be valid and reachable, otherwise an error will be returned
func NewSpeedcamsScraper(client *http.Client, baseRequestURL string) (*SpeedcamsScraper, error) {
	// Ping the base request URL to check if it's reachable
	_, err := client.Get(baseRequestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create speedcams scraper: %w", err)
	}

	return &SpeedcamsScraper{client: client, baseRequestURL: baseRequestURL}, nil
}

// Gets the latest speedcams data link from the website
// Returns an error if the request fails or if the link is not found
func (ss *SpeedcamsScraper) getLatestSpeedcamsLink() (string, error) {
	monthName := date.GetCurrentSpanishMonth()
	requestURL := fmt.Sprintf("%s/?s=radar+%s", ss.baseRequestURL, monthName)

	res, err := ss.client.Get(requestURL)
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

	speedcamsDataLink := doc.Find("a[rel='bookmark']").First().AttrOr("href", "")
	if speedcamsDataLink == "" {
		return "", errors.New("failed to get latest speedcams link: no speedcams data link found")
	}

	return speedcamsDataLink, nil
}

// Gets the speedcams data rows from the given link
// Returns an error if the request fails or if the rows are not found
func (ss *SpeedcamsScraper) getSpeedcamsRowsFromLink(link string) ([]SpeedcamsRow, error) {
	res, err := ss.client.Get(link)
	if err != nil {
		return nil, fmt.Errorf("failed to get speedcams rows from link: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get speedcams rows from link: request failed with status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get speedcams rows from link: %w", err)
	}

	// Get speedcams data table rows
	rows := make([]SpeedcamsRow, 0, 62)
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		// Skip the first row (table header)
		if i == 0 {
			return
		}

		row := SpeedcamsRow{}

		// Day
		dayStr := strings.TrimSpace(strings.ReplaceAll(s.Find("td").Eq(0).Text(), "\\xa0", ""))
		row.Day, err = strconv.Atoi(dayStr)
		if err != nil && len(rows) > 0 {
			// If the day is not found, use the previous row's day
			// (this happens when the day is not specified in the table)
			previousRow := rows[len(rows)-1]
			row.Day = previousRow.Day
		}

		// Shift
		shiftStr := strings.TrimSpace(s.Find("td").Eq(1).Text())
		switch shiftStr {
		case "mañana":
			row.Shift = Morning
		case "tarde":
			row.Shift = Afternoon
		}

		// Streets
		streetsStr := strings.TrimSpace(s.Find("td").Eq(2).Text())
		row.Streets = filterEmptyStrings(strings.Split(streetsStr, "\n"))

		// Speed limits
		speedLimitsStr := strings.TrimSpace(s.Find("td").Eq(3).Text())
		row.SpeedLimits, _ = stringSliceToIntSlice(filterEmptyStrings(strings.Split(speedLimitsStr, "\n")))

		rows = append(rows, row)
	})

	return rows, nil
}

// Gets today's speedcams data from the website
// Returns an error if the request fails or if the data is not found
func (ss *SpeedcamsScraper) GetTodaysSpeedcamsData() (SpeedcamsDayData, error) {
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
			todaysSpeedcamsData.Morning = appendSpeedcams(todaysSpeedcamsData.Morning, row.Streets, row.SpeedLimits)
		case Afternoon:
			todaysSpeedcamsData.Afternoon = appendSpeedcams(todaysSpeedcamsData.Afternoon, row.Streets, row.SpeedLimits)
		default:
			log.Warn("Unexpected shift value: ", row.Shift)
		}
	}

	return todaysSpeedcamsData, nil
}
