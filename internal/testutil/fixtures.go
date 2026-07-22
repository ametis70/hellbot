package testutil

import (
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

// Base timestamps matching real API data (season 159)
var (
	T0 = time.Unix(1784505880, 0).UTC() // campaign snapshot time
)

// Defend event fixtures — all use event_id 5080 from real API data

func DefendEventActive() *domain.DefendEvent {
	return &domain.DefendEvent{
		Season:         159,
		ID:             5080,
		StartTime:      time.Unix(1784501941, 0).UTC(),
		EndTime:        time.Unix(1784674741, 0).UTC(),
		Region:         0,
		Enemy:          domain.EnemyIlluminate,
		PointsMax:      31602,
		Points:         486,
		Status:         domain.EventStatusActive,
		PlayersAtStart: 0,
	}
}

func DefendEventFailed() *domain.DefendEvent {
	e := DefendEventActive()
	e.Status = domain.EventStatusFail
	return e
}

func DefendEventSucceeded() *domain.DefendEvent {
	e := DefendEventActive()
	e.Status = domain.EventStatusSuccess
	return e
}

func DefendEventNewActive() *domain.DefendEvent {
	return &domain.DefendEvent{
		Season:    159,
		ID:        5081,
		StartTime: time.Unix(1784674742, 0).UTC(),
		EndTime:   time.Unix(1784847542, 0).UTC(),
		Region:    1,
		Enemy:     domain.EnemyCyborg,
		PointsMax: 25000,
		Points:    0,
		Status:    domain.EventStatusActive,
	}
}

// Attack event fixtures — based on real API data (event_ids 922, 923, 924)

func AttackEventActive() domain.AttackEvent {
	return domain.AttackEvent{
		Season:         159,
		ID:             924,
		StartTime:      time.Unix(1784291521, 0).UTC(),
		EndTime:        time.Unix(1784456642, 0).UTC(),
		Enemy:          domain.EnemyCyborg,
		PointsMax:      31576,
		Points:         0,
		Status:         domain.EventStatusActive,
		PlayersAtStart: 184,
		MaxEventID:     924,
	}
}

func AttackEventSucceeded() domain.AttackEvent {
	e := AttackEventActive()
	e.Status = domain.EventStatusSuccess
	e.Points = e.PointsMax
	return e
}

func AttackEventFailed() domain.AttackEvent {
	e := AttackEventActive()
	e.ID = 923
	e.Status = domain.EventStatusFail
	e.Enemy = domain.EnemyIlluminate
	e.PlayersAtStart = 302
	e.MaxEventID = 923
	return e
}

// Faction status fixtures

func FactionStatuses() []domain.FactionStatus {
	return []domain.FactionStatus{
		{
			Season:            159,
			Points:            280970,
			PointsTaken:       604351,
			PointsMax:         280970,
			Status:            domain.FactionStatusDefeated,
			IntroductionOrder: 2,
		},
		{
			Season:            159,
			Points:            351,
			PointsTaken:       484634,
			PointsMax:         325480,
			Status:            domain.FactionStatusActive,
			IntroductionOrder: 1,
		},
		{
			Season:            159,
			Points:            194728,
			PointsTaken:       295878,
			PointsMax:         202300,
			Status:            domain.FactionStatusActive,
			IntroductionOrder: 0,
		},
	}
}

// Campaign status builders

func CampaignWithActiveDefend() *domain.CampaignStatus {
	return &domain.CampaignStatus{
		Time:           T0,
		FactionsStatus: FactionStatuses(),
		DefendEvent:    DefendEventActive(),
		AttackEvents:   []domain.AttackEvent{},
		Statistics:     []domain.Statistics{},
	}
}

func CampaignWithFailedDefend() *domain.CampaignStatus {
	return &domain.CampaignStatus{
		Time:           T0.Add(time.Hour),
		FactionsStatus: FactionStatuses(),
		DefendEvent:    DefendEventFailed(),
		AttackEvents:   []domain.AttackEvent{},
		Statistics:     []domain.Statistics{},
	}
}

func CampaignWithSucceededDefend() *domain.CampaignStatus {
	return &domain.CampaignStatus{
		Time:           T0.Add(time.Hour),
		FactionsStatus: FactionStatuses(),
		DefendEvent:    DefendEventSucceeded(),
		AttackEvents:   []domain.AttackEvent{},
		Statistics:     []domain.Statistics{},
	}
}

func CampaignWithNoDefend() *domain.CampaignStatus {
	return &domain.CampaignStatus{
		Time:           T0,
		FactionsStatus: FactionStatuses(),
		DefendEvent:    nil,
		AttackEvents:   []domain.AttackEvent{},
		Statistics:     []domain.Statistics{},
	}
}

func CampaignWithActiveAttack() *domain.CampaignStatus {
	return &domain.CampaignStatus{
		Time:           T0,
		FactionsStatus: FactionStatuses(),
		DefendEvent:    nil,
		AttackEvents:   []domain.AttackEvent{AttackEventActive()},
		Statistics:     []domain.Statistics{},
	}
}

func CampaignWithEndedAttack() *domain.CampaignStatus {
	return &domain.CampaignStatus{
		Time:           T0.Add(time.Hour),
		FactionsStatus: FactionStatuses(),
		DefendEvent:    nil,
		AttackEvents:   []domain.AttackEvent{AttackEventSucceeded()},
		Statistics:     []domain.Statistics{},
	}
}
