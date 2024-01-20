package speedcamsscraper

// Enum representing the shift of the day in which the speedcams are active
type Shift int

const (
	Morning Shift = iota
	Afternoon
)

// Row of the speedcams table in the website
type SpeedcamsRow struct {
	Day         int
	Shift       Shift
	Streets     []string
	SpeedLimits []int
}
