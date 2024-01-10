package utils

import "time"

var spanishMonths = [...]string{
	"enero",
	"febrero",
	"marzo",
	"abril",
	"mayo",
	"junio",
	"julio",
	"agosto",
	"septiembre",
	"octubre",
	"noviembre",
	"diciembre",
}

// Returns the name of the current month in Spanish
func GetCurrentSpanishMonth() string {
	monthNum := int(time.Now().Month()) - 1
	return spanishMonths[monthNum]
}
