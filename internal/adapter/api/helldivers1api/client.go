package helldivers1api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

type Options struct {
	BaseURL     string
	Timeout     time.Duration
	InsecureTLS bool
}

func DefaultOptions() Options {
	return Options{
		BaseURL:     "https://api.helldiversgame.com/1.0/",
		Timeout:     10 * time.Second,
		InsecureTLS: true,
	}
}

type Client struct {
	opts   Options
	http   *http.Client
	logger *slog.Logger
}

func New(opts Options, logger *slog.Logger) *Client {
	if logger == nil {
		panic("logger is required")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.InsecureTLS},
	}

	return &Client{
		opts:   opts,
		logger: logger,
		http: &http.Client{
			Timeout:   opts.Timeout,
			Transport: transport,
		},
	}
}

type apiFactionStatus struct {
	Season            int    `json:"season"`
	Points            int    `json:"points"`
	PointsTaken       int    `json:"points_taken"`
	PointsMax         int    `json:"points_max"`
	Status            string `json:"status"`
	IntroductionOrder int    `json:"introduction_order"`
}

type apiDefendEvent struct {
	Season         int    `json:"season"`
	ID             int    `json:"event_id"`
	StartTime      int    `json:"start_time"`
	EndTime        int    `json:"end_time"`
	Region         int    `json:"region"`
	Enemy          int    `json:"enemy"`
	PointsMax      int    `json:"points_max"`
	Points         int    `json:"points"`
	Status         string `json:"status"`
	PlayersAtStart int    `json:"players_at_start"`
}

type apiAttackEvent struct {
	Season         int    `json:"season"`
	ID             int    `json:"event_id"`
	StartTime      int    `json:"start_time"`
	EndTime        int    `json:"end_time"`
	Enemy          int    `json:"enemy"`
	PointsMax      int    `json:"points_max"`
	Points         int    `json:"points"`
	Status         string `json:"status"`
	PlayersAtStart int    `json:"players_at_start"`
	MaxEventID     int    `json:"max_event_id"`
}

type apiStatistics struct {
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

type apiCampaignStatus struct {
	Time          int                `json:"time"`
	ErrorCode     int                `json:"error_code"`
	FactionStatus []apiFactionStatus `json:"campaign_status"`
	DefendEvent   apiDefendEvent     `json:"defend_event"`
	AttackEvents  []apiAttackEvent   `json:"attack_events"`
	Statistics    []apiStatistics    `json:"statistics"`
}

func toDomainFactionStatus(a apiFactionStatus, enemy domain.Enemy) domain.FactionStatus {
	return domain.FactionStatus{
		Enemy:             enemy,
		Season:            a.Season,
		Points:            a.Points,
		PointsTaken:       a.PointsTaken,
		PointsMax:         a.PointsMax,
		Status:            domain.FactionStatusKind(a.Status),
		IntroductionOrder: a.IntroductionOrder,
	}
}

func toDomainDefendEvent(a apiDefendEvent) *domain.DefendEvent {
	return &domain.DefendEvent{
		Season:         a.Season,
		ID:             a.ID,
		StartTime:      time.Unix(int64(a.StartTime), 0).UTC(),
		EndTime:        time.Unix(int64(a.EndTime), 0).UTC(),
		Region:         a.Region,
		Enemy:          domain.Enemy(a.Enemy),
		PointsMax:      a.PointsMax,
		Points:         a.Points,
		Status:         domain.EventStatusKind(a.Status),
		PlayersAtStart: a.PlayersAtStart,
	}
}

func toDomainAttackEvent(a apiAttackEvent) domain.AttackEvent {
	return domain.AttackEvent{
		Season:         a.Season,
		ID:             a.ID,
		StartTime:      time.Unix(int64(a.StartTime), 0).UTC(),
		EndTime:        time.Unix(int64(a.EndTime), 0).UTC(),
		Enemy:          domain.Enemy(a.Enemy),
		PointsMax:      a.PointsMax,
		Points:         a.Points,
		Status:         domain.EventStatusKind(a.Status),
		PlayersAtStart: a.PlayersAtStart,
		MaxEventID:     a.MaxEventID,
	}
}

func toDomainStatistics(a apiStatistics) domain.Statistics {
	return domain.Statistics{
		Season:                 a.Season,
		SeasonDuration:         a.SeasonDuration,
		Enemy:                  domain.Enemy(a.Enemy),
		Players:                a.Players,
		TotalUniquePlayers:     a.TotalUniquePlayers,
		Missions:               a.Missions,
		SuccessfulMissions:     a.SuccessfulMissions,
		TotalMissionDifficulty: a.TotalMissionDifficulty,
		CompletedPlanets:       a.CompletedPlanets,
		DefendEvents:           a.DefendEvents,
		SuccessfulDefendEvents: a.SuccessfulDefendEvents,
		AttackEvents:           a.AttackEvents,
		SuccessfulAttackEvents: a.SuccessfulAttackEvents,
		Deaths:                 a.Deaths,
		Kills:                  a.Kills,
		Accidentals:            a.Accidentals,
		Shots:                  a.Shots,
		Hits:                   a.Hits,
	}
}

func toDomainCampaign(a apiCampaignStatus) *domain.CampaignStatus {
	factions := make([]domain.FactionStatus, len(a.FactionStatus))
	for i, f := range a.FactionStatus {
		factions[i] = toDomainFactionStatus(f, domain.Enemy(i))
	}

	attacks := make([]domain.AttackEvent, len(a.AttackEvents))
	for i, e := range a.AttackEvents {
		attacks[i] = toDomainAttackEvent(e)
	}

	stats := make([]domain.Statistics, len(a.Statistics))
	for i, s := range a.Statistics {
		stats[i] = toDomainStatistics(s)
	}

	var defendEvent *domain.DefendEvent
	if a.DefendEvent.ID != 0 {
		defendEvent = toDomainDefendEvent(a.DefendEvent)
	}

	return &domain.CampaignStatus{
		Time:           time.Unix(int64(a.Time), 0).UTC(),
		FactionsStatus: factions,
		DefendEvent:    defendEvent,
		AttackEvents:   attacks,
		Statistics:     stats,
	}
}

func (c *Client) FetchCampaign() (*domain.CampaignStatus, error) {
	resp, err := c.http.PostForm(c.opts.BaseURL, url.Values{
		"action": {"get_campaign_status"},
	})
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var apiResp apiCampaignStatus
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.ErrorCode != 0 {
		return nil, fmt.Errorf("api error code %d", apiResp.ErrorCode)
	}

	if len(apiResp.FactionStatus) > 0 {
		c.logger.Info("fetched campaign status", "season", apiResp.FactionStatus[0].Season)
	} else {
		c.logger.Info("fetched campaign status")
	}

	return toDomainCampaign(apiResp), nil
}
