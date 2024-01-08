package main

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

func GetThisSpanishMonth() string {
	monthNum := int(time.Now().Month())
	return spanishMonths[monthNum]
}
