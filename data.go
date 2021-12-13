package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

type FactionStatus struct {
	Time              int    `json:"-" db:"time"`
	Season            int    `json:"season" db:"season"`
	Points            int    `json:"points" db:"points"`
	PointsTaken       int    `json:"points_taken" db:"points_taken"`
	PointsMax         int    `json:"points_max" db:"points_max"`
	Status            string `json:"status" db:"status"`
	IntroductionOrder int    `json:"introduction_order" db:"introduction_order"`
}

type DefendEvent struct {
	Time      int    `json:"-" db:"time"`
	Season    int    `json:"season" db:"season"`
	ID        int    `json:"event_id" db:"event_id"`
	StartTime int    `json:"start_time" db:"start_time"`
	EndTime   int    `json:"end_time" db:"end_time"`
	Enemy     int    `json:"enemy" db:"enemy"`
	PointsMax int    `json:"points_max" db:"points_max"`
	Points    int    `json:"points" db:"points"`
	Status    string `json:"status" db:"status"`
	Region    int    `json:"region" db:"region"`
}

type AttackEvent struct {
	Time           int    `json:"-" db:"time"`
	Season         int    `json:"season" db:"season"`
	ID             int    `json:"event_id" db:"event_id"`
	StartTime      int    `json:"start_time" db:"start_time"`
	EndTime        int    `json:"end_time" db:"end_time"`
	Enemy          int    `json:"enemy" db:"enemy"`
	PointsMax      int    `json:"points_max" db:"points_max"`
	Points         int    `json:"points" db:"points"`
	Status         string `json:"status" db:"status"`
	PlayersAtStart int    `json:"players_at_start" db:"players_at_start"`
	MaxEventId     int    `json:"max_event_id" db:"max_event_id"`
}

type Statistics struct {
	Time                   int `json:"-" db:"time"`
	Season                 int `json:"season" db:"season"`
	SeasonDuration         int `json:"season_duration" db:"season_duration"`
	Enemy                  int `json:"enemy" db:"enemy"`
	Players                int `json:"players" db:"players"`
	TotalUniquePlayers     int `json:"total_unique_players" db:"total_unique_players"`
	Missions               int `json:"missions" db:"missions"`
	SuccessfulMissions     int `json:"successful_missions" db:"successful_missions"`
	TotalMissionDifficulty int `json:"total_mission_difficulty" db:"total_mission_difficulty"`
	CompletedPlanets       int `json:"completed_planets" db:"completed_planets"`
	DefendEvents           int `json:"defend_events" db:"defend_events"`
	SuccessfulDefendEvents int `json:"successful_defend_events" db:"successful_defend_events"`
	AttackEvents           int `json:"attack_events" db:"attack_events"`
	SuccessfulAttackEvents int `json:"successful_attack_events" db:"successful_attack_events"`
	Deaths                 int `json:"deaths" db:"deaths"`
	Kills                  int `json:"kills" db:"kills"`
	Accidentals            int `json:"accidentals" db:"accidentals"`
	Shots                  int `json:"shots" db:"shots"`
	Hits                   int `json:"hits" db:"hits"`
}

type CampaignStatus struct {
	Time  int `db:"time"`
	Error int `db:"error"`
}

type Data struct {
	Time          int             `json:"time"`
	Error         int             `json:"error_code"`
	FactionStatus []FactionStatus `json:"campaign_status"`
	DefendEvent   DefendEvent     `json:"defend_event"`
	AttackEvents  []AttackEvent   `json:"attack_events"`
	Statistics    []Statistics    `json:"statistics"`
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
