package main

import (
	"fmt"
	"os"

	"github.com/dmarts05/leon-speedcams-go/internal/config"
	"github.com/dmarts05/leon-speedcams-go/internal/message"
	"github.com/dmarts05/leon-speedcams-go/internal/speedcamsscraper"
	"github.com/dmarts05/leon-speedcams-go/internal/timeoutclient"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

// Initialize the logger with the following configuration:
// - Log level: debug
// - Report caller: true
// - Log file: logs/app.log
// - Log rotation: true
func initLogger() {
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

	log.Info("Reading config file...")
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Config: ", conf)

	log.Info("Extracting today's speedcams data...")
	client := timeoutclient.NewTimeoutClient(conf.RequestTimeout)
	defer client.CloseIdleConnections()

	scraper, err := speedcamsscraper.NewSpeedcamsScraper(client, conf.BaseRequestURL)
	if err != nil {
		log.Fatal(err)
	}

	data, err := scraper.GetTodaysSpeedcamsData()
	if err != nil {
		log.Fatal(err)
	}

	msg := message.BuildSpeedcamsDayDataMessage(data, conf.MonitoredStreets)
	fmt.Print(msg)

	log.Info("Sending today's speedcams data to Telegram...")
	telegramSender := message.TelegramBotMessageSender{Client: client, Token: conf.TelegramBotToken, ChatID: conf.TelegramChatID}
	err = telegramSender.SendMessage(msg)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Done!")
}
