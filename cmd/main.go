package main

import (
	"fmt"

	"github.com/dmarts05/leon-speedcams-go/internal/config"
	"github.com/dmarts05/leon-speedcams-go/internal/message"
	"github.com/dmarts05/leon-speedcams-go/internal/speedcamsscraper"
	"github.com/dmarts05/leon-speedcams-go/internal/timeoutclient"
	log "github.com/sirupsen/logrus"
)

// Initializes the logger with debug level and caller reporting
func initLogger() {
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
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

	scraper := speedcamsscraper.SpeedcamsScraper{Client: client, BaseRequestURL: conf.BaseRequestURL}

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
