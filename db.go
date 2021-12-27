package main

import (
	"bytes"
	"html/template"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sqlx.DB
)

func OpenDatabase() *sqlx.DB {
	_, debug := os.LookupEnv("DEBUG")
	if debug {
		os.Remove("db.sqlite")
	}

	db = sqlx.MustConnect("sqlite3", "file:db.sqlite?fk=true")
	return db
}

func InitDatabase() {
	foreignKeySchema := `
    time INTEGER,
    FOREIGN KEY(time) REFERENCES campaign_status(time) ON UPDATE CASCADE ON DELETE CASCADE`

	eventSchema := `
    season INTEGER,
    event_id INTEGER,
    start_time INTEGER,
    end_time INTEGER,
    enemy INTEGER,
    points_max INTEGER,
    points INTEGER,
    status TEXT`

	s := `
CREATE TABLE IF NOT EXISTS campaign_status (
    time INTEGER UNIQUE NOT NULL PRIMARY KEY,
    error integer not null
);

CREATE TABLE IF NOT EXISTS faction_status (
    season INTEGER,
    points INTEGER,
    points_taken INTEGER,
    points_max INTEGER,
    status TEXT,
    introduction_order INTEGER,{{.fk}}
);

CREATE TABLE IF NOT EXISTS attack_events ({{.event}},
    players_at_start INTEGER,
    max_event_id INTEGER,{{.fk}}
);

CREATE TABLE IF NOT EXISTS defend_events ({{.event}},
    region INTEGER,{{.fk}}
);

CREATE TABLE IF NOT EXISTS statistics (
    season INTEGER,
    season_duration INTEGER, 
    enemy INTEGER,
    players INTEGER,
    total_unique_players INTEGER,
    missions INTEGER,
    successful_missions INTEGER,
    total_mission_difficulty INTEGER,
    completed_planets INTEGER,
    defend_events INTEGER,
    successful_defend_events INTEGER,
    attack_events INTEGER,
    successful_attack_events INTEGER,
    deaths INTEGER,
    kills INTEGER,
    accidentals INTEGER,
    shots INTEGER,
    hits INTEGER,{{.fk}}
);

CREATE TABLE IF NOT EXISTS ongoing_events (
    id INTEGER PRIMARY_KEY,
    event_type TEXT
)
`

	var sqlStmt bytes.Buffer
	t := template.Must(template.New("s").Parse(s))
	t.Execute(&sqlStmt, map[string]interface{}{"fk": foreignKeySchema, "event": eventSchema})
	db.MustExec(sqlStmt.String(), foreignKeySchema, eventSchema)
}

func StoreData(data *Data) error {
	tx := db.MustBegin()
	tx.MustExec("INSERT INTO campaign_status (time, error) VALUES (?, ?)", data.Time, data.Error)

	for _, faction := range data.FactionStatus {
		tx.MustExec(
			`INSERT INTO faction_status 
            (season, points, points_taken, points_max, status, introduction_order, time)
            VALUES (?, ?, ?, ?, ?, ?, ?)`,
			faction.Season, faction.Points, faction.PointsTaken, faction.PointsMax, faction.Status,
			faction.IntroductionOrder, data.Time)
	}

	for _, attack := range data.AttackEvents {
		tx.MustExec(
			`INSERT INTO attack_events
            (season, event_id, start_time, end_time, enemy, points_max, points, status, players_at_start, max_event_id, time)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			attack.Season, attack.ID, attack.StartTime, attack.EndTime, attack.Enemy, attack.PointsMax,
			attack.Points, attack.Status, attack.PlayersAtStart, attack.MaxEventId, data.Time,
		)
	}

	defend := data.DefendEvent
	tx.MustExec(
		`INSERT INTO defend_events 
            (season, event_id, start_time, end_time, enemy, points_max, points, status, region, time)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		defend.Season, defend.ID, defend.StartTime, defend.EndTime, defend.Enemy, defend.PointsMax,
		defend.Points, defend.Status, defend.Region, data.Time,
	)

	for _, statistics := range data.Statistics {
		tx.MustExec(
			`INSERT INTO statistics 
            (season, season_duration, enemy, players, total_unique_players, missions, successful_missions, total_mission_difficulty,
            completed_planets, defend_events, successful_defend_events, attack_events, successful_attack_events, deaths, kills,
            accidentals, shots, hits, time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			statistics.Season, statistics.SeasonDuration, statistics.Enemy, statistics.Players, statistics.TotalUniquePlayers,
			statistics.Missions, statistics.SuccessfulMissions, statistics.TotalMissionDifficulty, statistics.CompletedPlanets,
			statistics.DefendEvents, statistics.SuccessfulDefendEvents, statistics.AttackEvents, statistics.SuccessfulAttackEvents,
			statistics.Deaths, statistics.Kills, statistics.Accidentals, statistics.Shots, statistics.Hits, data.Time,
		)
	}

	if err := tx.Commit(); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func GetDefendEventById(id int) (*DefendEvent, error) {
	defendEvent := DefendEvent{}
	err := db.Get(&defendEvent, "SELECT * FROM defend_events WHERE id=?", id)

	if err != nil {
		log.Print("Failed to retrieve defend_event data")
		return nil, err
	}

	return &defendEvent, nil
}

func GetLatestData() (*Data, error) {
	var err error
	campaignStatus := CampaignStatus{}
	err = db.Get(&campaignStatus, "SELECT * FROM campaign_status ORDER BY time DESC LIMIT 1")
	if err != nil {
		return nil, err
	}

	factionStatus := []FactionStatus{}
	defendEvent := DefendEvent{}
	attackEvents := []AttackEvent{}
	statistics := []Statistics{}

	tx := db.MustBegin()
	err = tx.Select(&factionStatus, "SELECT * FROM faction_status WHERE time=? ORDER BY introduction_order ASC", campaignStatus.Time)
	if err != nil {
		log.Print("Failed to retrieve faction_status data")
		return nil, err
	}
	err = tx.Get(&defendEvent, "SELECT * FROM defend_events WHERE time=?", campaignStatus.Time)
	if err != nil {
		log.Print("Failed to retrieve defend_event data")
		return nil, err
	}
	err = tx.Select(&attackEvents, "SELECT * FROM attack_events WHERE time=? ORDER BY enemy ASC", campaignStatus.Time)
	if err != nil {
		log.Print("Failed to retrieve attack_events data")
		return nil, err
	}
	err = tx.Select(&statistics, "SELECT * FROM statistics WHERE time=? ORDER BY enemy ASC", campaignStatus.Time)
	if err != nil {
		log.Print("Failed to retrieve statistics data")
		return nil, err
	}
	tx.Commit()

	data := Data{
		Time:          campaignStatus.Time,
		Error:         campaignStatus.Error,
		FactionStatus: factionStatus,
		DefendEvent:   defendEvent,
		AttackEvents:  attackEvents,
		Statistics:    statistics,
	}

	return &data, nil
}
