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

	pData, err := GetLatestData()
	StoreData(data)

	if err != nil {
		log.Println("There's no previous data to compare, skipping iteration")
		return
	}

	channel := os.Getenv("CHANNEL_ID")

	if (pData.DefendEvent.Status != data.DefendEvent.Status) ||
		(pData.DefendEvent.ID != data.DefendEvent.ID) {
		p := pData.DefendEvent
		n := data.DefendEvent

		switch n.Status {
		case "active":
			if p.Status == "failure" || p.Status == "success" || p.ID != p.ID {
				msg := fmt.Sprintf("New defend event against %v in region %v\nStart Time: %v\nEnd time: %v\nID: %v",
					n.Enemy, n.Region, time.Unix(int64(n.StartTime), 0).UTC(), time.Unix(int64(n.EndTime), 0).UTC(), n.ID)
				dg.ChannelMessageSend(channel, msg)
			}

		case "failure":
			if p.Status == "active" && p.ID == n.ID {
				msg := fmt.Sprintf("We failed! the %v have taken back region %v\nID: %v",
					n.Enemy, n.Region, n.ID)
				dg.ChannelMessageSend(channel, msg)
			}

		case "success":
			if p.Status == "active" && p.ID == n.ID {
				msg := fmt.Sprintf("We did it! Super Earth has conquered region %v against %v\nID: %v",
					n.Region, n.Enemy, n.ID)
				dg.ChannelMessageSend(channel, msg)
			}

		}
	}
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
