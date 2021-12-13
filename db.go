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
    FOREIGN KEY(time) REFERENCES campaign_status(time) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT pk PRIMARY KEY(time)`

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
