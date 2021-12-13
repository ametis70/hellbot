package main

import (
	"bytes"
	"database/sql"
	"html/template"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDatabase() {
	os.Remove("db.sqlite")

	db, err := sql.Open("sqlite3", "file:db.sqlite?fk=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
CREATE TABLE campaign_status (
    time INTEGER UNIQUE NOT NULL PRIMARY KEY,
    error integer not null
);

CREATE TABLE faction_status (
    season INTEGER,
    points INTEGER,
    points_taken INTEGER,
    points_max INTEGER,
    status TEXT,
    introduction_order INTEGER,{{.fk}}
);

CREATE TABLE attack_events ({{.event}},
    players_at_start INTEGER,
    max_event_id INTEGER,{{.fk}}
);

CREATE TABLE defend_events ({{.event}},
    region INTEGER,{{.fk}}
);

CREATE TABLE statistics (
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
`

	var sqlStmt bytes.Buffer
	t := template.Must(template.New("s").Parse(s))
	t.Execute(&sqlStmt, map[string]interface{}{"fk": foreignKeySchema, "event": eventSchema})
	_, err = db.Exec(sqlStmt.String(), foreignKeySchema, eventSchema)

	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func StoreData(data *Data) {
	db, err := sql.Open("sqlite3", "file:db.sqlite?fk=true")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	tx, err := db.Begin()

	if err != nil {
		log.Fatal(err)
	}

	_, campaignStatusErr := tx.Exec("INSERT INTO campaign_status (time, error) VALUES (?, ?)", data.Time, data.Error)

	if campaignStatusErr != nil {
		_ = tx.Rollback()
		log.Fatal(campaignStatusErr)
	}

	for _, faction := range data.FactionStatus {
		_, factionStatusErr := tx.Exec(
			`INSERT INTO faction_status 
            (season, points, points_taken, points_max, status, time)
            VALUES (?, ?, ?, ?, ?, ?)`,
			faction.Season, faction.Points, faction.PointsTaken, faction.PointsMax, faction.Status, data.Time)

		if factionStatusErr != nil {
			_ = tx.Rollback()
			log.Fatal(factionStatusErr)
		}
	}

	for _, attack := range data.AttackEvents {
		_, attackEventErr := tx.Exec(
			`INSERT INTO attack_events
            (season, event_id, start_time, end_time, enemy, points_max, points, status, players_at_start, max_event_id, time)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			attack.Season, attack.ID, attack.StartTime, attack.EndTime, attack.Enemy, attack.PointsMax,
			attack.Points, attack.Status, attack.PlayersAtStart, attack.MaxEventId, data.Time,
		)

		if attackEventErr != nil {
			_ = tx.Rollback()
			log.Fatal(attackEventErr)
		}
	}

	defend := data.DefendEvent
	_, defendEventErr := tx.Exec(
		`INSERT INTO defend_events 
            (season, event_id, start_time, end_time, enemy, points_max, points, status, region, time)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		defend.Season, defend.ID, defend.StartTime, defend.EndTime, defend.Enemy, defend.PointsMax,
		defend.Points, defend.Status, defend.Region, data.Time,
	)

	if defendEventErr != nil {
		_ = tx.Rollback()
		log.Fatal(defendEventErr)
	}

	for _, statistics := range data.Statistics {
		_, statisticsErr := tx.Exec(
			`INSERT INTO statistics 
    (season, season_duration, enemy, players, total_unique_players, missions, successful_missions, total_mission_difficulty, completed_planets, defend_events, successful_defend_events, attack_events, successful_attack_events, deaths, kills, accidentals, shots, hits, time)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			statistics.Season, statistics.SeasonDuration, statistics.Enemy, statistics.Players, statistics.TotalUniquePlayers,
			statistics.Missions, statistics.SuccessfulMissions, statistics.TotalMissionDifficulty, statistics.CompletedPlanets,
			statistics.DefendEvents, statistics.SuccessfulDefendEvents, statistics.AttackEvents, statistics.SuccessfulAttackEvents,
			statistics.Deaths, statistics.Kills, statistics.Accidentals, statistics.Shots, statistics.Hits, data.Time,
		)

		if statisticsErr != nil {
			_ = tx.Rollback()
			log.Fatal(statisticsErr)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
