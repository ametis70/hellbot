package main

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

type messageType uint8

const (
	eventNew messageType = iota
	eventSuccess
	eventFail
)

type OngoingEvent struct {
	ID        int    `db:"event_id"`
	EventType string `db:"event_type"`
}

func storeOngoingEvent(id int, eventType string) error {
	_, err := db.Exec(
		`INSERT INTO ongoing_events (event_id, event_type) VALUES (?, ?)`, id, eventType)

	return err
}

func removeOngoingEvent(id int, eventType string) error {
	_, err := db.Exec(
		`DELETE FROM ongoing_events WHERE event_id=? AND event_type=?`, id, eventType)

	return err
}

func getOngoingEvents() ([]*OngoingEvent, error) {
	ongoingEvents := []*OngoingEvent{}
	err := db.Select(&ongoingEvents, "SELECT * FROM ongoing_events")

	if err != nil {
		return nil, err
	}

	return ongoingEvents, nil
}

func getOngoingEvent(id int, eventType string) (*OngoingEvent, error) {
	ongoingEvents := OngoingEvent{}
	err := db.Get(&ongoingEvents,
		"SELECT * FROM ongoing_events WHERE event_id=? AND event_type=?", id, eventType)

	if err != nil {
		return nil, err
	}

	return &ongoingEvents, nil
}

func getDefendEventStatusFromSnapshot(season int, id int) (string, error) {
	data, err := FetchSnapshot(season)

	if err != nil {
		return "", err
	}

	var status string

	for _, e := range data.DefendEvents {
		if e.ID == id {
			status = e.Status
			break
		} else {
			logger.Errorw(
				"Defend ID not found in snapshot",
				zap.Int("season", season),
				zap.Int("id", e.ID),
			)
		}
	}
	return status, nil
}

func sendDefendMessage(mt messageType, data *DefendEvent) {
	discordChannel := os.Getenv("CHANNEL_ID")
	var msg string

	switch mt {
	case eventNew:
		msg = fmt.Sprintf(
			"New defend event against %v in region %v\nStart Time: %v\nEnd time: %v\nID: %v",
			data.Enemy,
			data.Region,
			time.Unix(int64(data.StartTime), 0).UTC(),
			time.Unix(int64(data.EndTime), 0).UTC(),
			data.ID,
		)

	case eventFail:
		msg = fmt.Sprintf(
			"We failed! the %v have taken back region %v\nID: %v",
			data.Enemy,
			data.Region,
			data.ID,
		)

	case eventSuccess:
		msg = fmt.Sprintf(
			"We did it! Super Earth has conquered region %v against %v\nID: %v",
			data.Region,
			data.Enemy,
			data.ID,
		)
	}

	_, err := dg.ChannelMessageSend(discordChannel, msg)

	if err != nil {
		logger.Error("An error has ocurred while sending message: ", err)
	}
}

func handleDefendEvent(data *DefendEvent) {
	storedEvent, _ := getOngoingEvent(data.ID, "defend")

	// There is an active event but it wasn't stored yet
	if storedEvent == nil {
		if data.Status == "active" {
			storeOngoingEvent(data.ID, "defend")
			sendDefendMessage(eventNew, data)
		}

		return
	}

	storedEventData, err := GetDefendEventById(storedEvent.ID)

	if err != nil {
		logger.Error("Error retrieving stored event data: ", err)
	}

	// Handle event with new id
	if storedEvent.ID != data.ID {
		snapshotStatus, _ := getDefendEventStatusFromSnapshot(
			storedEventData.Season,
			storedEventData.ID,
		)

		if storedEventData != nil && snapshotStatus != "" {
			switch snapshotStatus {
			case "success":
				sendDefendMessage(eventSuccess, storedEventData)
			case "fail":
				sendDefendMessage(eventFail, storedEventData)
			}
		}

		storeOngoingEvent(data.ID, "defend")
		sendDefendMessage(eventNew, data)

		// Remove old event
		removeOngoingEvent(storedEvent.ID, "defend")
	}
}

func HandleEvents(data *CampaignStatus) {
	handleDefendEvent(&data.DefendEvent)
}
