package speedcamsscraper

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGetSpeedcamsRowsFromBody_TableFormat(t *testing.T) {
	body := strings.NewReader(`
		<div class="entry-content">
			<table><tbody>
				<tr><td>Día</td><td>Turno</td><td>Calles</td><td>Velocidad</td></tr>
				<tr><td>11</td><td>mañana</td><td>José Aguado<br>LE-20, Navatejera</td><td>30<br>70</td></tr>
				<tr><td>&nbsp;</td><td>tarde</td><td>Avda. Europa</td><td>50 km/h</td></tr>
			</tbody></table>
		</div>`)

	rows, err := (SpeedcamsScraper{}).getSpeedcamsRowsFromBody(body)
	if err != nil {
		t.Fatalf("getSpeedcamsRowsFromBody() error = %v", err)
	}

	want := []SpeedcamsRow{
		{
			Day:   11,
			Shift: Morning,
			Speedcams: []Speedcam{
				{Street: "José Aguado", SpeedLimit: 30},
				{Street: "LE-20, Navatejera", SpeedLimit: 70},
			},
		},
		{
			Day:       11,
			Shift:     Afternoon,
			Speedcams: []Speedcam{{Street: "Avda. Europa", SpeedLimit: 50}},
		},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("getSpeedcamsRowsFromBody() = %#v, want %#v", rows, want)
	}
}

func TestGetSpeedcamsRowsFromBody_ProseFormat(t *testing.T) {
	body := strings.NewReader(`
		<div class="entry-content">
			<h3>Día 11 de agosto</h3>
			<p><strong>Mañana:</strong> Pº Salamanca (<strong>30 km/h</strong>); LE-20, Navatejera (<strong>70 km/h</strong>).</p>
			<p><strong>Tarde:</strong> Avda. Padre Isla, prioridad peatonal (<strong>10 km/h</strong>); Avda. Europa (<strong>50 km/h</strong>).</p>
			<h3>Día 12 de agosto</h3>
			<p><strong>Mañana:</strong> Santo Tirso (<strong>30 km/h</strong>).</p>
		</div>`)

	rows, err := (SpeedcamsScraper{}).getSpeedcamsRowsFromBody(body)
	if err != nil {
		t.Fatalf("getSpeedcamsRowsFromBody() error = %v", err)
	}

	want := []SpeedcamsRow{
		{
			Day:   11,
			Shift: Morning,
			Speedcams: []Speedcam{
				{Street: "Pº Salamanca", SpeedLimit: 30},
				{Street: "LE-20, Navatejera", SpeedLimit: 70},
			},
		},
		{
			Day:   11,
			Shift: Afternoon,
			Speedcams: []Speedcam{
				{Street: "Avda. Padre Isla, prioridad peatonal", SpeedLimit: 10},
				{Street: "Avda. Europa", SpeedLimit: 50},
			},
		},
		{
			Day:       12,
			Shift:     Morning,
			Speedcams: []Speedcam{{Street: "Santo Tirso", SpeedLimit: 30}},
		},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("getSpeedcamsRowsFromBody() = %#v, want %#v", rows, want)
	}
}

func TestParseSpeedcamsRows_PrefersTableFormat(t *testing.T) {
	doc := documentFromHTML(t, `
		<div class="entry-content">
			<table><tr><td>8</td><td>mañana</td><td>Table Street</td><td>30</td></tr></table>
			<h3>Día 8 de agosto</h3>
			<p>Mañana: Prose Street (50 km/h).</p>
		</div>`)

	rows, err := parseSpeedcamsRows(doc)
	if err != nil {
		t.Fatalf("parseSpeedcamsRows() error = %v", err)
	}
	if got := rows[0].Speedcams[0].Street; got != "Table Street" {
		t.Fatalf("parseSpeedcamsRows() street = %q, want table parser result", got)
	}
}

func TestParseSpeedcamsRows_RejectsMalformedRecognizedTableWithoutFallback(t *testing.T) {
	doc := documentFromHTML(t, `
		<div class="entry-content">
			<table><tr><td>8</td><td>mañana</td><td>First Street<br>Second Street</td><td>30</td></tr></table>
			<h3>Día 8 de agosto</h3>
			<p>Mañana: Prose Street (50 km/h).</p>
		</div>`)

	_, err := parseSpeedcamsRows(doc)
	if err == nil || !strings.Contains(err.Error(), "2 streets and 1 speed limits") {
		t.Fatalf("parseSpeedcamsRows() error = %v, want mismatched table columns", err)
	}
}

func TestParseSpeedcamsRows_UnsupportedFormat(t *testing.T) {
	doc := documentFromHTML(t, `<div class="entry-content"><p>No radar data here.</p></div>`)

	_, err := parseSpeedcamsRows(doc)
	if !errors.Is(err, errUnsupportedStructure) {
		t.Fatalf("parseSpeedcamsRows() error = %v, want errUnsupportedStructure", err)
	}
}

func TestTableRowsParser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		wantError string
	}{
		{
			name:      "invalid day",
			html:      `<table><tr><td>32</td><td>mañana</td><td>Street</td><td>30</td></tr></table>`,
			wantError: "outside the valid range",
		},
		{
			name:      "unknown shift",
			html:      `<table><tr><td>1</td><td>noche</td><td>Street</td><td>30</td></tr></table>`,
			wantError: "unknown shift",
		},
		{
			name:      "invalid limit",
			html:      `<table><tr><td>1</td><td>tarde</td><td>Street</td><td>fast</td></tr></table>`,
			wantError: "invalid speed limit",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := (tableRowsParser{}).Parse(documentFromHTML(t, test.html))
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Parse() error = %v, want error containing %q", err, test.wantError)
			}
		})
	}
}

func TestProseRowsParser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		wantError string
	}{
		{
			name:      "invalid day",
			html:      `<div class="entry-content"><h3>Día 32 de agosto</h3><p>Mañana: Street (30 km/h).</p></div>`,
			wantError: "outside the valid range",
		},
		{
			name:      "unknown shift",
			html:      `<div class="entry-content"><h3>Día 1 de agosto</h3><p>Mañana: First Street (30 km/h).</p><p>Noche: Second Street (30 km/h).</p></div>`,
			wantError: `unknown shift "Noche"`,
		},
		{
			name:      "malformed entry",
			html:      `<div class="entry-content"><h3>Día 1 de agosto</h3><p>Mañana: Street (fast).</p></div>`,
			wantError: "is malformed",
		},
		{
			name: "duplicate shift",
			html: `<div class="entry-content">
				<h3>Día 1 de agosto</h3>
				<p>Mañana: First Street (30 km/h).</p>
				<p>Mañana: Second Street (50 km/h).</p>
			</div>`,
			wantError: "duplicate morning shift",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := (proseRowsParser{}).Parse(documentFromHTML(t, test.html))
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Parse() error = %v, want error containing %q", err, test.wantError)
			}
		})
	}
}

func documentFromHTML(t *testing.T, value string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(value))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() error = %v", err)
	}
	return doc
}
