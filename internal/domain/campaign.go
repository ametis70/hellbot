package domain

import (
	"fmt"
	"time"
)

type Enemy int

const (
	EnemyCyborg     Enemy = 0
	EnemyIlluminate Enemy = 1
	EnemyBug        Enemy = 2
)

func (e Enemy) String() string {
	switch e {
	case EnemyCyborg:
		return "Cyborg"
	case EnemyIlluminate:
		return "Illuminate"
	case EnemyBug:
		return "Bug"
	default:
		return fmt.Sprintf("Unknown(%d)", int(e))
	}
}

type EventStatusKind string

const (
	EventStatusActive  EventStatusKind = "active"
	EventStatusSuccess EventStatusKind = "success"
	EventStatusFail    EventStatusKind = "fail"
)

type FactionStatusKind string

const (
	FactionStatusActive   FactionStatusKind = "active"
	FactionStatusDefeated FactionStatusKind = "defeated"
	FactionStatusHidden   FactionStatusKind = "hidden"
)

type FactionStatus struct {
	Season            int
	Points            int
	PointsTaken       int
	PointsMax         int
	Status            FactionStatusKind
	IntroductionOrder int
}

type DefendEvent struct {
	Season         int
	ID             int
	StartTime      time.Time
	EndTime        time.Time
	Region         int
	Enemy          Enemy
	PointsMax      int
	Points         int
	Status         EventStatusKind
	PlayersAtStart int
}

type AttackEvent struct {
	Season         int
	ID             int
	StartTime      time.Time
	EndTime        time.Time
	Enemy          Enemy
	PointsMax      int
	Points         int
	Status         EventStatusKind
	PlayersAtStart int
	MaxEventID     int
}

type Statistics struct {
	Season                 int
	SeasonDuration         int
	Enemy                  Enemy
	Players                int
	TotalUniquePlayers     int
	Missions               int
	SuccessfulMissions     int
	TotalMissionDifficulty int
	CompletedPlanets       int
	DefendEvents           int
	SuccessfulDefendEvents int
	AttackEvents           int
	SuccessfulAttackEvents int
	Deaths                 int
	Kills                  int
	Accidentals            int
	Shots                  int
	Hits                   int
}

type CampaignStatus struct {
	Time           time.Time
	FactionsStatus []FactionStatus
	DefendEvent    *DefendEvent
	AttackEvents   []AttackEvent
	Statistics     []Statistics
}
