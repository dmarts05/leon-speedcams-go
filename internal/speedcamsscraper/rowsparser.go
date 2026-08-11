package speedcamsscraper

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	errUnsupportedStructure = errors.New("unsupported speedcams structure")
	dayHeadingPattern       = regexp.MustCompile(`(?i)^d[ií]a\s+(\d{1,2})(?:\s+de\b.*)?$`)
	proseShiftPattern       = regexp.MustCompile(`(?i)^\s*(mañana|tarde)\s*:\s*(.+?)\s*$`)
	proseLabelPattern       = regexp.MustCompile(`^\s*([^:]+)\s*:\s*(.+?)\s*$`)
	proseSpeedcamPattern    = regexp.MustCompile(`(?i)^(.+?)\s*\(\s*(\d+)\s*km\s*/\s*h\s*\)\s*\.?$`)
	tableLimitPattern       = regexp.MustCompile(`(?i)^(\d+)(?:\s*km\s*/\s*h)?$`)
)

// Strategy for parsing speedcams rows from a supported document structure
type speedcamsRowsParser interface {
	Parse(*goquery.Document) ([]SpeedcamsRow, error)
}

// Parses speedcams rows using the first strategy that supports the document structure
func parseSpeedcamsRows(doc *goquery.Document) ([]SpeedcamsRow, error) {
	parsers := []speedcamsRowsParser{
		tableRowsParser{},
		proseRowsParser{},
	}

	for _, parser := range parsers {
		rows, err := parser.Parse(doc)
		if errors.Is(err, errUnsupportedStructure) {
			continue
		}
		return rows, err
	}

	return nil, errUnsupportedStructure
}

// Parses speedcams rows from the regular table structure
type tableRowsParser struct{}

// Parses and validates all speedcams rows in the first radar table
func (tableRowsParser) Parse(doc *goquery.Document) ([]SpeedcamsRow, error) {
	table := firstRadarTable(doc)
	if table == nil {
		return nil, errUnsupportedStructure
	}

	rows := make([]SpeedcamsRow, 0, 62)
	var parseErr error
	table.Find("tr").EachWithBreak(func(index int, selection *goquery.Selection) bool {
		cells := selection.Find("td")
		if cells.Length() == 0 {
			return true
		}
		if cells.Length() < 4 {
			parseErr = fmt.Errorf("table row %d has %d columns, expected at least 4", index+1, cells.Length())
			return false
		}

		dayText := normalizeText(cells.Eq(0).Text())
		shiftText := normalizeText(cells.Eq(1).Text())
		if len(rows) == 0 && isTableHeader(dayText, shiftText) {
			return true
		}

		day, err := parseTableDay(dayText, rows)
		if err != nil {
			parseErr = fmt.Errorf("table row %d: %w", index+1, err)
			return false
		}
		shift, err := parseShift(shiftText)
		if err != nil {
			parseErr = fmt.Errorf("table row %d: %w", index+1, err)
			return false
		}

		streets := selectionLines(cells.Eq(2))
		limits := selectionLines(cells.Eq(3))
		if len(streets) == 0 {
			parseErr = fmt.Errorf("table row %d has no streets", index+1)
			return false
		}
		if len(streets) != len(limits) {
			parseErr = fmt.Errorf("table row %d has %d streets and %d speed limits", index+1, len(streets), len(limits))
			return false
		}

		speedcams := make([]Speedcam, 0, len(streets))
		for itemIndex, street := range streets {
			limit, err := parseTableLimit(limits[itemIndex])
			if err != nil {
				parseErr = fmt.Errorf("table row %d speed limit %d: %w", index+1, itemIndex+1, err)
				return false
			}
			speedcams = append(speedcams, Speedcam{Street: street, SpeedLimit: limit})
		}

		rows = append(rows, SpeedcamsRow{Day: day, Shift: shift, Speedcams: speedcams})
		return true
	})

	if parseErr != nil {
		return nil, parseErr
	}
	if len(rows) == 0 {
		return nil, errors.New("radar table contains no data rows")
	}

	return rows, nil
}

// Finds the first table that contains radar data columns
func firstRadarTable(doc *goquery.Document) *goquery.Selection {
	var table *goquery.Selection
	doc.Find(".entry-content table, table").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		if selection.Find("tr").FilterFunction(func(_ int, row *goquery.Selection) bool {
			return row.Find("td").Length() >= 4
		}).Length() == 0 {
			return true
		}
		table = selection
		return false
	})
	return table
}

// Checks whether the given day and shift cells belong to a table header
func isTableHeader(day, shift string) bool {
	day = strings.ToLower(day)
	shift = strings.ToLower(shift)
	return (strings.Contains(day, "día") || strings.Contains(day, "dia")) && strings.Contains(shift, "turno")
}

// Parses a table day, using the previous row's day when the cell is empty
func parseTableDay(value string, previousRows []SpeedcamsRow) (int, error) {
	if value == "" {
		if len(previousRows) == 0 {
			return 0, errors.New("day is empty without a previous row")
		}
		return previousRows[len(previousRows)-1].Day, nil
	}

	day, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid day %q", value)
	}
	if err := validateDay(day); err != nil {
		return 0, err
	}
	return day, nil
}

// Parses a numeric table speed limit with an optional km/h suffix
func parseTableLimit(value string) (int, error) {
	matches := tableLimitPattern.FindStringSubmatch(normalizeText(value))
	if len(matches) != 2 {
		return 0, fmt.Errorf("invalid speed limit %q", value)
	}
	limit, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid speed limit %q: %w", value, err)
	}
	return limit, nil
}

// Gets the non-empty text lines from a table cell, including lines separated by br elements
func selectionLines(selection *goquery.Selection) []string {
	selection = selection.Clone()
	selection.Find("br").ReplaceWithHtml("\n")

	lines := strings.Split(selection.Text(), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line = normalizeText(line); line != "" {
			result = append(result, line)
		}
	}
	return result
}

// Parses speedcams rows from day headings followed by shift paragraphs
type proseRowsParser struct{}

// Parses and validates all speedcams rows in the heading and paragraph structure
func (proseRowsParser) Parse(doc *goquery.Document) ([]SpeedcamsRow, error) {
	headings := doc.Find(".entry-content h3").FilterFunction(func(_ int, selection *goquery.Selection) bool {
		return dayHeadingPattern.MatchString(normalizeText(selection.Text()))
	})
	if headings.Length() == 0 {
		return nil, errUnsupportedStructure
	}

	rows := make([]SpeedcamsRow, 0, headings.Length()*2)
	seen := make(map[[2]int]struct{}, headings.Length()*2)
	var parseErr error
	headings.EachWithBreak(func(_ int, heading *goquery.Selection) bool {
		day, err := parseHeadingDay(heading.Text())
		if err != nil {
			parseErr = err
			return false
		}

		rowsBeforeHeading := len(rows)
		for sibling := heading.Next(); sibling.Length() > 0 && !isHeading(sibling); sibling = sibling.Next() {
			if !sibling.Is("p") {
				continue
			}
			shift, speedcams, recognized, err := parseProseShift(sibling.Text())
			if err != nil {
				parseErr = fmt.Errorf("day %d: %w", day, err)
				return false
			}
			if !recognized {
				continue
			}

			key := [2]int{day, int(shift)}
			if _, exists := seen[key]; exists {
				parseErr = fmt.Errorf("day %d contains duplicate %s shift", day, shiftName(shift))
				return false
			}
			seen[key] = struct{}{}
			rows = append(rows, SpeedcamsRow{Day: day, Shift: shift, Speedcams: speedcams})
		}
		if parseErr != nil {
			return false
		}
		if len(rows) == rowsBeforeHeading {
			parseErr = fmt.Errorf("day %d contains no recognized shifts", day)
			return false
		}
		return true
	})

	if parseErr != nil {
		return nil, parseErr
	}
	return rows, nil
}

// Parses and validates the day number from a heading
func parseHeadingDay(value string) (int, error) {
	matches := dayHeadingPattern.FindStringSubmatch(normalizeText(value))
	if len(matches) != 2 {
		return 0, fmt.Errorf("invalid day heading %q", value)
	}
	day, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid day heading %q: %w", value, err)
	}
	if err := validateDay(day); err != nil {
		return 0, err
	}
	return day, nil
}

// Parses a shift paragraph containing semicolon-separated speedcams
func parseProseShift(value string) (Shift, []Speedcam, bool, error) {
	value = normalizeText(value)
	matches := proseShiftPattern.FindStringSubmatch(value)
	if len(matches) != 3 {
		labelParts := proseLabelPattern.FindStringSubmatch(value)
		if len(labelParts) == 3 && strings.Contains(strings.ToLower(labelParts[2]), "km/h") {
			return 0, nil, true, fmt.Errorf("unknown shift %q", strings.TrimSpace(labelParts[1]))
		}
		return 0, nil, false, nil
	}
	shift, err := parseShift(matches[1])
	if err != nil {
		return 0, nil, true, err
	}

	items := strings.Split(matches[2], ";")
	speedcams := make([]Speedcam, 0, len(items))
	for index, item := range items {
		item = normalizeText(item)
		if item == "" {
			return 0, nil, true, fmt.Errorf("%s entry %d is empty", shiftName(shift), index+1)
		}
		parts := proseSpeedcamPattern.FindStringSubmatch(item)
		if len(parts) != 3 {
			return 0, nil, true, fmt.Errorf("%s entry %d is malformed: %q", shiftName(shift), index+1, item)
		}
		limit, err := strconv.Atoi(parts[2])
		if err != nil {
			return 0, nil, true, fmt.Errorf("%s entry %d has invalid speed limit: %w", shiftName(shift), index+1, err)
		}
		speedcams = append(speedcams, Speedcam{Street: strings.TrimSpace(parts[1]), SpeedLimit: limit})
	}
	if len(speedcams) == 0 {
		return 0, nil, true, fmt.Errorf("%s shift contains no speedcams", shiftName(shift))
	}
	return shift, speedcams, true, nil
}

// Converts a Spanish shift name to its corresponding Shift value
func parseShift(value string) (Shift, error) {
	switch strings.ToLower(normalizeText(value)) {
	case "mañana":
		return Morning, nil
	case "tarde":
		return Afternoon, nil
	default:
		return 0, fmt.Errorf("unknown shift %q", value)
	}
}

// Gets the English name of a Shift value for error messages
func shiftName(shift Shift) string {
	if shift == Morning {
		return "morning"
	}
	return "afternoon"
}

// Checks whether a day is within the valid calendar range
func validateDay(day int) error {
	if day < 1 || day > 31 {
		return fmt.Errorf("day %d is outside the valid range", day)
	}
	return nil
}

// Checks whether the selection is an HTML heading element
func isHeading(selection *goquery.Selection) bool {
	switch goquery.NodeName(selection) {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	default:
		return false
	}
}

// Normalizes non-breaking spaces and repeated whitespace in text
func normalizeText(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(value, "\u00a0", " ")), " ")
}
