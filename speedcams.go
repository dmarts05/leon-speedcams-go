package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"

	log "github.com/sirupsen/logrus"
)

const (
	baseRequestURL = "https://www.ahoraleon.com"
)

func getLatestSpeedcamsDataLink() (string, error) {
	monthName := GetThisSpanishMonth()
	requestURL := fmt.Sprintf("%s/?s=radar+%s", baseRequestURL, monthName)

	resp, err := http.Get(requestURL)
	if err != nil {
		log.Error("Error while getting latest speedcams data link: ", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New("request failed with status code: " + strconv.Itoa(resp.StatusCode))
		log.Error("Request failed with status code: ", resp.StatusCode)
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Error("Error while parsing response body: ", err)
		return "", err
	}

	speedcamsDataLink := doc.Find("a[rel='bookmark']").First().AttrOr("href", "")
	if speedcamsDataLink == "" {
		err = errors.New("no speedcams data link found")
		log.Error("Error while getting latest speedcams data link: ", err)
		return "", err
	}

	return speedcamsDataLink, nil
}

func ExtractSpeedcamsData() {
	panic("Not implemented")
}
