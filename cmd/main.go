package main

import (
	"fmt"
	"os"

	"github.com/dmarts05/leon-speedcams-go/internal/speedcamsscraper"
	"github.com/dmarts05/leon-speedcams-go/internal/timeoutclient"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

const (
	requestTimeout = 30
	baseRequestURL = "https://www.ahoraleon.com"
)

// Initialize the logger with the following configuration:
// - Log level: debug
// - Report caller: true
// - Log file: logs/app.log
// - Log rotation: true
func initLogger() {
	// TODO: add option in config file to set log level
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)

	// Configure log rotation for file logging
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		// Use lfshook to write logs to file with rotation
		log.AddHook(lfshook.NewHook(
			lfshook.WriterMap{
				log.DebugLevel: logFile,
				log.InfoLevel:  logFile,
				log.WarnLevel:  logFile,
				log.ErrorLevel: logFile,
				log.FatalLevel: logFile,
				log.PanicLevel: logFile,
			},
			&log.JSONFormatter{},
		))
	} else {
		log.Warn("Failed to log to file, using default stderr")
	}
}

// Show a welcome message with the app name
func showWelcomeMessage() {
	fmt.Println(`
██╗     ███████╗ ██████╗ ███╗   ██╗                                        
██║     ██╔════╝██╔═══██╗████╗  ██║                                        
██║     █████╗  ██║   ██║██╔██╗ ██║                                        
██║     ██╔══╝  ██║   ██║██║╚██╗██║                                        
███████╗███████╗╚██████╔╝██║ ╚████║                                        
╚══════╝╚══════╝ ╚═════╝ ╚═╝  ╚═══╝                                        
                                                                           
███████╗██████╗ ███████╗███████╗██████╗  ██████╗ █████╗ ███╗   ███╗███████╗
██╔════╝██╔══██╗██╔════╝██╔════╝██╔══██╗██╔════╝██╔══██╗████╗ ████║██╔════╝
███████╗██████╔╝█████╗  █████╗  ██║  ██║██║     ███████║██╔████╔██║███████╗
╚════██║██╔═══╝ ██╔══╝  ██╔══╝  ██║  ██║██║     ██╔══██║██║╚██╔╝██║╚════██║
███████║██║     ███████╗███████╗██████╔╝╚██████╗██║  ██║██║ ╚═╝ ██║███████║
╚══════╝╚═╝     ╚══════╝╚══════╝╚═════╝  ╚═════╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝`)
}

func main() {
	initLogger()

	showWelcomeMessage()

	// log.Info("Reading config file...")
	// TODO: Read .toml config file

	log.Info("Extracting today's speedcams data...")
	client := timeoutclient.NewTimeoutClient(requestTimeout)
	defer client.CloseIdleConnections()

	scraper, err := speedcamsscraper.NewSpeedcamsScraper(client, baseRequestURL)
	if err != nil {
		log.Fatal(err)
	}
	data, err := scraper.GetTodaysSpeedcamsData()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(data.String())

	// log.Info("Sending today's speedcams data to Telegram...")
}
