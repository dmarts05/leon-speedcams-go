package speedcamsscraper

// Enum representing the shift of the day in which the speedcams are active
type Shift int

const (
	Morning Shift = iota
	Afternoon
)

// Speedcams data for one day and shift in the website.
type SpeedcamsRow struct {
	Day       int
	Shift     Shift
	Speedcams []Speedcam
}
