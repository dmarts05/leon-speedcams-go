package main

import (
	"fmt"
	"os"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

func initLogger() {
	// TODO: add option in config file to set log level
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)

	// Configure log rotation for file logging
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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

	log.Info("Reading config file...")
	// TODO: Read .toml config file

	log.Info("Extracting today's speedcams data...")
}
