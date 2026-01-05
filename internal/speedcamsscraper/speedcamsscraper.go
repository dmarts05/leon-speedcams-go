// Module for scraping speedcams data from the given website (https://ahoraleon.com)
package speedcamsscraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dmarts05/leon-speedcams-go/internal/spanishdate"

	log "github.com/sirupsen/logrus"
)

// Regex to extract street name and speed limit, e.g., "Street Name (30)"
var streetLimitRegex = regexp.MustCompile(`(.*?)\s*\((\d+)\)`)

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

// Appends the given streets and speed limits to the given speedcams slice
func appendSpeedcams(existing []Speedcam, streets []string, limits []int) []Speedcam {
	for i, street := range streets {
		existing = append(existing, Speedcam{Street: street, SpeedLimit: limits[i]})
	}
	return existing
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

// Parses a raw text line containing streets and limits (e.g. "Street A (30), Street B (50)")
func (ss SpeedcamsScraper) parseStreetsAndLimits(text string) ([]string, []int) {
	var streets []string
	var limits []int

	// Clean up labels safely (ReplaceAll is safer than TrimPrefix for unstructured text)
	text = strings.ReplaceAll(text, "Mañana:", "")
	text = strings.ReplaceAll(text, "Tarde:", "")
	text = strings.TrimSpace(text)

	// Split by comma
	segments := strings.Split(text, ",")

	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		matches := streetLimitRegex.FindStringSubmatch(segment)
		if len(matches) == 3 {
			street := strings.TrimSpace(matches[1])
			limit, err := strconv.Atoi(matches[2])
			if err == nil {
				streets = append(streets, street)
				limits = append(limits, limit)
			}
		}
	}
	return streets, limits
}

// Gets the speedcams data rows from the given body
// Returns an error if the rows are not found
func (ss SpeedcamsScraper) getSpeedcamsRowsFromBody(body io.ReadCloser) ([]SpeedcamsRow, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to get speedcams rows from website body: %w", err)
	}

	var rows []SpeedcamsRow

	// Iterate over H3 headers which contain "Día X" text
	doc.Find(".entry-content h3").Each(func(i int, s *goquery.Selection) {
		dayText := strings.TrimSpace(s.Text())
		if !strings.HasPrefix(dayText, "Día") {
			return
		}

		// Extract Day Number
		dayParts := strings.Split(dayText, " ")
		if len(dayParts) < 2 {
			return
		}
		day, err := strconv.Atoi(dayParts[1])
		if err != nil {
			return
		}

		// The content is in the paragraph immediately following the H3
		contentP := s.Next()
		if !contentP.Is("p") {
			return
		}

		fullText := contentP.Text()

		// Locate the "Tarde:" marker to split Morning and Afternoon
		// We use "Tarde:" as a separator because "Mañana:" is usually at the start.
		tardeIndex := strings.Index(fullText, "Tarde:")

		var morningText, afternoonText string

		if tardeIndex != -1 {
			morningText = fullText[:tardeIndex]
			afternoonText = fullText[tardeIndex:]
		} else {
			// Handle cases where only one shift might be present
			if strings.Contains(fullText, "Mañana") {
				morningText = fullText
			} else if strings.Contains(fullText, "Tarde") {
				afternoonText = fullText
			}
		}

		// Parse Morning
		if morningText != "" {
			streets, limits := ss.parseStreetsAndLimits(morningText)
			if len(streets) > 0 {
				rows = append(rows, SpeedcamsRow{
					Day:         day,
					Shift:       Morning,
					Streets:     streets,
					SpeedLimits: limits,
				})
			}
		}

		// Parse Afternoon
		if afternoonText != "" {
			streets, limits := ss.parseStreetsAndLimits(afternoonText)
			if len(streets) > 0 {
				rows = append(rows, SpeedcamsRow{
					Day:         day,
					Shift:       Afternoon,
					Streets:     streets,
					SpeedLimits: limits,
				})
			}
		}
	})

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

// Gets today speedcams data from the website
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
			todaysSpeedcamsData.Morning = appendSpeedcams(todaysSpeedcamsData.Morning, row.Streets, row.SpeedLimits)
		case Afternoon:
			todaysSpeedcamsData.Afternoon = appendSpeedcams(todaysSpeedcamsData.Afternoon, row.Streets, row.SpeedLimits)
		default:
			log.Warn("Unexpected shift value: ", row.Shift)
		}
	}

	return todaysSpeedcamsData, nil
}
