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

type Shift int

const (
	Morning Shift = iota
	Afternoon
)

type speedcamsRow struct {
	Day         int
	Shift       Shift
	Streets     []string
	SpeedLimits []int
}

type Speedcam struct {
	Street     string
	SpeedLimit int
}

type SpeedcamsDayData struct {
	Day       int
	Morning   []Speedcam
	Afternoon []Speedcam
}

func getLatestSpeedcamsLink(client *http.Client, baseRequestURL string) (string, error) {
	monthName := utils.GetThisSpanishMonth()
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

func getSpeedcamsRowsFromLink(client *http.Client, link string) ([]speedcamsRow, error) {
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
	rows := make([]speedcamsRow, 0, 62)
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		row := speedcamsRow{}

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
					row.Shift = Morning
				case "tarde":
					row.Shift = Afternoon
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

func GetTodaysSpeedcamsData(client *http.Client, baseRequestURL string) (SpeedcamsDayData, error) {
	speedcamsDataLink, err := getLatestSpeedcamsLink(client, baseRequestURL)
	if err != nil {
		return SpeedcamsDayData{}, err
	}

	log.Debug("Latest speedcams data link: ", speedcamsDataLink)

	speedcamRows, err := getSpeedcamsRowsFromLink(client, speedcamsDataLink)
	if err != nil {
		return SpeedcamsDayData{}, err
	}

	// Filter speedcams data rows to get today's data
	today := int(time.Now().Day())
	todayRows := make([]speedcamsRow, 0, 2)
	for _, row := range speedcamRows {
		if row.Day == today {
			todayRows = append(todayRows, row)
		}
	}

	// Get today's speedcams data
	todaysSpeedcamsData := SpeedcamsDayData{Day: today}
	for _, row := range todayRows {
		switch row.Shift {
		case Morning:
			for i, street := range row.Streets {
				todaysSpeedcamsData.Morning = append(todaysSpeedcamsData.Morning, Speedcam{street, row.SpeedLimits[i]})
			}
		case Afternoon:
			for i, street := range row.Streets {
				todaysSpeedcamsData.Afternoon = append(todaysSpeedcamsData.Afternoon, Speedcam{street, row.SpeedLimits[i]})
			}
		default:
			log.Warn("Unexpected shift value: ", row.Shift)
		}
	}

	return todaysSpeedcamsData, nil
}
