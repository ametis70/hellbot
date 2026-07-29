package mock

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

// Fetcher is a port.Fetcher implementation that plays back a scripted sequence
// of CampaignStatus values, one per call. When the sequence is exhausted it
// cancels the provided context so the poller shuts down gracefully.
type Fetcher struct {
	mu     sync.Mutex
	frames []*domain.CampaignStatus
	idx    int
	cancel context.CancelFunc
	logger *slog.Logger
}

// New creates a Fetcher loaded with the built-in WarScenario. cancel is called
// once all frames have been served so the poller exits cleanly.
func New(cancel context.CancelFunc, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		frames: warScenario(),
		cancel: cancel,
		logger: logger,
	}
}

// FetchCampaign returns the next frame in the scenario. Once all frames have
// been served it cancels the context and returns the last frame so the final
// poll completes normally before the poller loop exits.
func (f *Fetcher) FetchCampaign() (*domain.CampaignStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	total := len(f.frames)

	if f.idx >= total {
		// Already exhausted — context should already be cancelled.
		return f.frames[total-1], nil
	}

	frame := f.frames[f.idx]
	f.logger.Info("mock: serving frame", "index", f.idx+1, "total", total)
	f.idx++

	if f.idx >= total {
		f.logger.Info("mock: scenario exhausted, stopping")
		f.cancel()
	}

	return frame, nil
}

// warScenario returns the scripted war sequence as domain values.
//
// Sequence (9 frames):
//
//  1. Idle — first poll, no events; poller stores baseline, no diff
//  2. Attack starts (active)
//  3. Attack still active
//  4. Attack succeeded
//  5. Idle
//  6. Defend starts (active)
//  7. Defend still active
//  8. Defend succeeded; all factions defeated
//  9. New season — war won
//
// Expected notifications: attack started, attack succeeded, defend started,
// defend succeeded, war won.
func warScenario() []*domain.CampaignStatus {
	const (
		season   = 50
		attackID = 100
		defendID = 200
	)

	t0 := time.Now().UTC().Truncate(time.Second)

	activeFactionsS50 := []domain.FactionStatus{
		{Enemy: domain.EnemyBug, Season: season, Points: 300000, PointsMax: 300000, Status: domain.FactionStatusActive, IntroductionOrder: 2},
		{Enemy: domain.EnemyCyborg, Season: season, Points: 400000, PointsMax: 400000, Status: domain.FactionStatusActive, IntroductionOrder: 1},
		{Enemy: domain.EnemyIlluminate, Season: season, Points: 200000, PointsMax: 200000, Status: domain.FactionStatusActive, IntroductionOrder: 0},
	}

	defeatedFactionsS50 := []domain.FactionStatus{
		{Enemy: domain.EnemyBug, Season: season, Points: 300000, PointsMax: 300000, Status: domain.FactionStatusDefeated, IntroductionOrder: 2},
		{Enemy: domain.EnemyCyborg, Season: season, Points: 400000, PointsMax: 400000, Status: domain.FactionStatusDefeated, IntroductionOrder: 1},
		{Enemy: domain.EnemyIlluminate, Season: season, Points: 200000, PointsMax: 200000, Status: domain.FactionStatusDefeated, IntroductionOrder: 0},
	}

	newSeasonFactions := []domain.FactionStatus{
		{Enemy: domain.EnemyBug, Season: season + 1, Points: 0, PointsMax: 300000, Status: domain.FactionStatusActive, IntroductionOrder: 2},
		{Enemy: domain.EnemyCyborg, Season: season + 1, Points: 0, PointsMax: 400000, Status: domain.FactionStatusActive, IntroductionOrder: 1},
		{Enemy: domain.EnemyIlluminate, Season: season + 1, Points: 0, PointsMax: 200000, Status: domain.FactionStatusActive, IntroductionOrder: 0},
	}

	attackActive := domain.AttackEvent{
		Season: season, ID: attackID,
		StartTime: t0.Add(60 * time.Second), EndTime: t0.Add(time.Hour),
		Enemy: domain.EnemyCyborg, PointsMax: 50000, Points: 0,
		Status: domain.EventStatusActive, MaxEventID: attackID,
	}
	attackSucceeded := attackActive
	attackSucceeded.Points = attackActive.PointsMax
	attackSucceeded.Status = domain.EventStatusSuccess

	defendActive := &domain.DefendEvent{
		Season: season, ID: defendID,
		StartTime: t0.Add(5 * time.Minute), EndTime: t0.Add(2 * time.Hour),
		Region: 3, Enemy: domain.EnemyIlluminate, PointsMax: 30000, Points: 0,
		Status: domain.EventStatusActive,
	}
	defendSucceeded := *defendActive
	defendSucceeded.Points = defendActive.PointsMax
	defendSucceeded.Status = domain.EventStatusSuccess

	idle := func(ts time.Time, factions []domain.FactionStatus) *domain.CampaignStatus {
		return &domain.CampaignStatus{
			Time:           ts,
			FactionsStatus: factions,
			AttackEvents:   []domain.AttackEvent{},
		}
	}

	return []*domain.CampaignStatus{
		// 1. Idle
		idle(t0, activeFactionsS50),
		// 2. Attack starts
		{Time: t0.Add(60 * time.Second), FactionsStatus: activeFactionsS50, AttackEvents: []domain.AttackEvent{attackActive}},
		// 3. Attack still active
		{Time: t0.Add(2 * time.Minute), FactionsStatus: activeFactionsS50, AttackEvents: []domain.AttackEvent{attackActive}},
		// 4. Attack succeeded
		{Time: t0.Add(3 * time.Minute), FactionsStatus: activeFactionsS50, AttackEvents: []domain.AttackEvent{attackSucceeded}},
		// 5. Idle
		idle(t0.Add(4*time.Minute), activeFactionsS50),
		// 6. Defend starts
		{Time: t0.Add(5 * time.Minute), FactionsStatus: activeFactionsS50, DefendEvent: defendActive, AttackEvents: []domain.AttackEvent{}},
		// 7. Defend still active
		{Time: t0.Add(6 * time.Minute), FactionsStatus: activeFactionsS50, DefendEvent: defendActive, AttackEvents: []domain.AttackEvent{}},
		// 8. Defend succeeded; all factions defeated
		{Time: t0.Add(7 * time.Minute), FactionsStatus: defeatedFactionsS50, DefendEvent: &defendSucceeded, AttackEvents: []domain.AttackEvent{}},
		// 9. New season — war won
		idle(t0.Add(8*time.Minute), newSeasonFactions),
	}
}
