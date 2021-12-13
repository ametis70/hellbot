package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

type FactionStatus struct {
	Season            int    `json:"season"`
	Points            int    `json:"points"`
	PointsTaken       int    `json:"points_taken"`
	PointsMax         int    `json:"points_max"`
	Status            string `json:"status"`
	IntroductionOrder int    `json:"introduction_order"`
}

type Event struct {
	Season    int    `json:"season"`
	ID        int    `json:"event_id"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Enemy     int    `json:"enemy"`
	PointsMax int    `json:"points_max"`
	Points    int    `json:"points"`
	Status    string `json:"status"`
}

type DefendEvent struct {
	Event
	Region int `json:"region"`
}

type AttackEvent struct {
	Event
	PlayersAtStart int `json:"players_at_start"`
	MaxEventId     int `json:"max_event_id"`
}

type Statistics struct {
	Season                 int `json:"season"`
	SeasonDuration         int `json:"season_duration"`
	Enemy                  int `json:"enemy"`
	Players                int `json:"players"`
	TotalUniquePlayers     int `json:"total_unique_players"`
	Missions               int `json:"missions"`
	SuccessfulMissions     int `json:"successful_missions"`
	TotalMissionDifficulty int `json:"total_mission_difficulty"`
	CompletedPlanets       int `json:"completed_planets"`
	DefendEvents           int `json:"defend_events"`
	SuccessfulDefendEvents int `json:"successful_defend_events"`
	AttackEvents           int `json:"attack_events"`
	SuccessfulAttackEvents int `json:"successful_attack_events"`
	Deaths                 int `json:"deaths"`
	Kills                  int `json:"kills"`
	Accidentals            int `json:"accidentals"`
	Shots                  int `json:"shots"`
	Hits                   int `json:"hits"`
}

type Data struct {
	Time          int              `json:"time"`
	Error         int              `json:"error_code"`
	FactionStatus [3]FactionStatus `json:"campaign_status"`
	DefendEvent   DefendEvent      `json:"defend_event"`
	AttackEvents  [3]AttackEvent   `json:"attack_events"`
	Statistics    [3]Statistics    `json:"statistics"`
}

func FetchData() (*Data, error) {
	data := Data{}

	// Workaround for self-signed certificate
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Get data
	endpoint := "https://api.helldiversgame.com/1.0/"

	formData := url.Values{
		"action": {"get_campaign_status"},
	}

	resp, httpErr := client.PostForm(endpoint, formData)

	if httpErr != nil {
		log.Fatalln(httpErr)
		return nil, httpErr
	}

	// Decode data
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	jsonErr := decoder.Decode(&data)

	if jsonErr != nil {
		log.Fatalln(jsonErr)
		return nil, jsonErr
	}

	return &data, nil
}
