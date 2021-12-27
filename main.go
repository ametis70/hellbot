package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

var (
	dg *discordgo.Session
)

func loop() {
	defer time.AfterFunc(time.Second*60, loop)

	data, err := FetchData()
	if err != nil {
		log.Println("Failed to fetch data, skipping iteration")
		return
	}

	StoreData(data)
	HandleEvents(data)
}

func main() {
	db = OpenDatabase()
	defer db.Close()

	InitDatabase()

	token := os.Getenv("TOKEN")

	var err error
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}
	defer dg.Close()

	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}

	loop()

	fmt.Println("Bot is now running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
