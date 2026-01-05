package main

import (
	"fmt"
	"time"

	"github.com/dmarts05/leon-speedcams-go/internal/config"
	"github.com/dmarts05/leon-speedcams-go/internal/message"
	"github.com/dmarts05/leon-speedcams-go/internal/speedcamsscraper"
	"github.com/dmarts05/leon-speedcams-go/internal/timeoutclient"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
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

// Run the speedcam job to extract today's speedcams data and send it to Telegram
func runSpeedcamJob(cfg config.Config) {
	log.Info("Extracting today's speedcams data...")

	client := timeoutclient.NewTimeoutClient(cfg.RequestTimeout)
	defer client.CloseIdleConnections()

	scraper := speedcamsscraper.SpeedcamsScraper{
		Client:         client,
		BaseRequestURL: cfg.BaseRequestURL,
	}

	data, err := scraper.GetTodaySpeedcamsData()
	if err != nil {
		log.Error("Failed to get speedcams data: ", err)
		return
	}

	msg := message.BuildSpeedcamsDayDataMessage(data, cfg.MonitoredStreets)
	fmt.Print(msg)

	log.Info("Sending today's speedcams data to Telegram...")
	telegramSender := message.TelegramBotMessageSender{
		Client: client, Token: cfg.TelegramBotToken, ChatID: cfg.TelegramChatID,
	}
	err = telegramSender.SendMessage(msg)
	if err != nil {
		log.Error("Failed to send Telegram message: ", err)
		return
	}

	log.Info("Job done!")
}

func main() {
	initLogger()
	showWelcomeMessage()

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Warn(".env file not found, skipping...")
	}

	log.Info("Reading config file...")
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Config: ", cfg)

	jobFunc := func() {
		runSpeedcamJob(cfg)
	}
	if cfg.EnableCron {
		s := gocron.NewScheduler(time.Local)
		_, err = s.Cron(cfg.Cron).Do(jobFunc)
		if err != nil {
			log.Fatal("Failed to schedule job with cron: ", err)
		}

		log.Infof("Scheduler enabled. Job will run with cron expression: %s", cfg.Cron)
		s.StartBlocking()
	} else {
		log.Info("Cron job is disabled by config. Running job once now...")
		jobFunc()
	}
}
