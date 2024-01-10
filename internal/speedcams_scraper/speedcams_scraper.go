package speedcams

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dmarts05/leon-speedcams-go/internal/utils"

	log "github.com/sirupsen/logrus"
)

// Get the latest speedcams data link from the website
// The link is the first link with the "bookmark" attribute
// The link is returned as a string or an error if the request fails or no link is found
func getLatestSpeedcamsLink(client *http.Client, baseRequestURL string) (string, error) {
	monthName := utils.GetCurrentSpanishMonth()
	requestURL := fmt.Sprintf("%s/?s=radar+%s", baseRequestURL, monthName)

	res, err := client.Get(requestURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = errors.New("request failed with status code: " + strconv.Itoa(res.StatusCode))
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	speedcamsDataLink := doc.Find("a[rel='bookmark']").First().AttrOr("href", "")
	if speedcamsDataLink == "" {
		err = errors.New("no speedcams data link found")
		return "", err
	}

	return speedcamsDataLink, nil
}

// Get the speedcams data rows from a specific link parsed with goquery
// Some rows may have empty values which is not a problem. We just skip them except for the day value (uses previous row's day)
// The rows are returned as a slice of SpeedcamsRow structs or an error if the request fails
func getSpeedcamsRowsFromLink(client *http.Client, link string) ([]utils.SpeedcamsRow, error) {
	res, err := client.Get(link)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = errors.New("request failed with status code: " + strconv.Itoa(res.StatusCode))
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// Get speedcams data table rows
	// Each day has 2 rows (morning and afternoon) and there are 31 days in a month (max) so we can allocate 62 rows
	rows := make([]utils.SpeedcamsRow, 0, 62)
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		row := utils.SpeedcamsRow{}

		// Skip first row (table header)
		if i == 0 {
			return
		}

		s.Find("td").Each(func(i int, s *goquery.Selection) {
			value := strings.TrimSpace(strings.ReplaceAll(s.Text(), "\\xa0", ""))

			// Fill row struct with data depending on column index
			switch i {
			// Day
			case 0:
				// If it's empty, use previous row's day
				if value == "" {
					row.Day = rows[len(rows)-1].Day
					return
				}

				day, err := strconv.Atoi(value)
				if err != nil {
					log.Warn("Error while converting day to int: ", err)
				} else {
					row.Day = day
				}
			// Shift
			case 1:
				switch value {
				case "mañana":
					row.Shift = utils.Morning
				case "tarde":
					row.Shift = utils.Afternoon
				default:
					log.Warn("Unexpected shift value: ", value)
				}
			// Streets
			case 2:
				// Split street names by new line and only append non-empty strings
				for _, street := range strings.Split(value, "\n") {
					if street != "" {
						row.Streets = append(row.Streets, street)
					}
				}
			// Speed limits
			case 3:
				// Split speed limits by new line and only append non-empty strings
				for _, speedLimit := range strings.Split(value, "\n") {
					if speedLimit != "" {
						speedLimit, err := strconv.Atoi(speedLimit)
						if err != nil {
							log.Warn("Error while converting speed limit to int: ", err)
						} else {
							row.SpeedLimits = append(row.SpeedLimits, speedLimit)
						}
					}
				}
			default:
				log.Warn("Unexpected column index: ", i)
			}
		})

		rows = append(rows, row)
	})

	return rows, nil
}

// Get today's speedcams data from the website using the given HTTP client and base request URL
// The data is returned as a SpeedcamsDayData struct or an error if the request fails
func GetTodaysSpeedcamsData(client *http.Client, baseRequestURL string) (utils.SpeedcamsDayData, error) {
	speedcamsDataLink, err := getLatestSpeedcamsLink(client, baseRequestURL)
	if err != nil {
		return utils.SpeedcamsDayData{}, err
	}

	log.Debug("Latest speedcams data link: ", speedcamsDataLink)

	speedcamRows, err := getSpeedcamsRowsFromLink(client, speedcamsDataLink)
	if err != nil {
		return utils.SpeedcamsDayData{}, err
	}

	// Filter speedcams data rows to get today's data
	today := time.Now()
	todayRows := make([]utils.SpeedcamsRow, 0, 2)
	for _, row := range speedcamRows {
		if row.Day == today.Day() {
			todayRows = append(todayRows, row)
		}
	}

	todaysSpeedcamsData := utils.SpeedcamsDayData{}
	todaysSpeedcamsData.Date = today

	// Get today's speedcams data
	for _, row := range todayRows {
		switch row.Shift {
		case utils.Morning:
			for i, street := range row.Streets {
				todaysSpeedcamsData.Morning = append(
					todaysSpeedcamsData.Morning,
					utils.Speedcam{Street: street, SpeedLimit: row.SpeedLimits[i]},
				)
			}
		case utils.Afternoon:
			for i, street := range row.Streets {
				todaysSpeedcamsData.Afternoon = append(
					todaysSpeedcamsData.Afternoon,
					utils.Speedcam{Street: street, SpeedLimit: row.SpeedLimits[i]},
				)
			}
		default:
			log.Warn("Unexpected shift value: ", row.Shift)
		}
	}

	return todaysSpeedcamsData, nil
}
