package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

var (
	dg     *discordgo.Session
	logger *zap.SugaredLogger
)

func loop() {
	defer time.AfterFunc(time.Second*60, loop)

	data, err := FetchData()
	if err != nil {
		logger.Warn("Failed to fetch data, skipping iteration")
		return
	}

	StoreData(data)
	HandleEvents(data)
}

func initLogger() {
	_logger, _ := zap.NewDevelopment()
	logger = _logger.Sugar()
}

func main() {
	initLogger()
	defer logger.Sync()

	db = OpenDatabase()
	defer db.Close()
	InitDatabase()

	var err error

	token := os.Getenv("TOKEN")
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		logger.Fatal("Error creating Discord session, ", err)
	}
	defer dg.Close()

	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		logger.Fatal("Error opening Discord connection, ", err)
		return
	}

	loop()

	logger.Info("Bot is now running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
