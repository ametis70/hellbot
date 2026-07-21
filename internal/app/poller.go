package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/port"
)

type Poller struct {
	fetcher   port.Fetcher
	campaigns port.CampaignStore
	events    port.EventStore
	notifiers []port.Notifier
	interval  time.Duration
	logger    *slog.Logger
}

func New(
	fetcher port.Fetcher,
	campaigns port.CampaignStore,
	events port.EventStore,
	notifiers []port.Notifier,
	interval time.Duration,
	logger *slog.Logger,
) *Poller {
	return &Poller{
		fetcher:   fetcher,
		campaigns: campaigns,
		events:    events,
		notifiers: notifiers,
		interval:  interval,
		logger:    logger,
	}
}

func (p *Poller) Run(ctx context.Context) error {
	p.poll()
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			p.poll()
		}
	}
}

func (p *Poller) poll() {
	current, err := p.fetcher.FetchCampaign()
	if err != nil {
		p.logger.Error("failed to fetch campaign", "error", err)
		return
	}

	previous, err := p.campaigns.LatestCampaign()
	if err != nil {
		p.logger.Warn("no previous campaign stored, skipping event detection")
	} else {
		p.handleEvents(current, previous)
	}

	if err := p.campaigns.SaveCampaign(current); err != nil {
		p.logger.Error("failed to save campaign", "error", err)
	}
}

func (p *Poller) notify(msg domain.EventMessage) {
	for _, n := range p.notifiers {
		if err := n.Notify(msg); err != nil {
			p.logger.Error("failed to send notification", "error", err)
		}
	}
}

func (p *Poller) handleEvents(current, previous *domain.CampaignStatus) {
	p.handleDefendEvent(current)
	p.handleAttackEvents(current)
}

func (p *Poller) handleDefendEvent(current *domain.CampaignStatus) {
	stored, err := p.events.ListOngoingEvents(domain.EventKindDefend)
	if err != nil {
		p.logger.Error("failed to list ongoing defend events", "error", err)
		return
	}

	// at most one defend event at a time
	var storedEvent *domain.OngoingEvent
	if len(stored) > 0 {
		storedEvent = stored[0]
	}

	if current.DefendEvent == nil && storedEvent == nil {
		return
	}

	if current.DefendEvent != nil && storedEvent == nil {
		p.events.SaveOngoingEvent(current.DefendEvent.ID, domain.EventKindDefend)
		p.notify(domain.EventMessage{
			Kind:        domain.EventKindDefend,
			Transition:  domain.EventTransitionStarted,
			DefendEvent: current.DefendEvent,
		})
		return
	}

	if current.DefendEvent == nil && storedEvent != nil {
		p.events.RemoveOngoingEvent(storedEvent.ID, domain.EventKindDefend)
		return
	}

	if current.DefendEvent.ID != storedEvent.ID {
		p.events.RemoveOngoingEvent(storedEvent.ID, domain.EventKindDefend)
		p.events.SaveOngoingEvent(current.DefendEvent.ID, domain.EventKindDefend)
		p.notify(domain.EventMessage{
			Kind:        domain.EventKindDefend,
			Transition:  domain.EventTransitionStarted,
			DefendEvent: current.DefendEvent,
		})
	}
}

func (p *Poller) handleAttackEvents(current *domain.CampaignStatus) {
	stored, err := p.events.ListOngoingEvents(domain.EventKindAttack)
	if err != nil {
		p.logger.Error("failed to list ongoing attack events", "error", err)
		return
	}

	// build map of current active attack IDs for O(1) lookup
	currentActive := make(map[int]domain.AttackEvent)
	for _, e := range current.AttackEvents {
		if e.Status == domain.EventStatusActive {
			currentActive[e.ID] = e
		}
	}

	// stored events not in current → ended
	for _, s := range stored {
		if _, stillActive := currentActive[s.ID]; !stillActive {
			p.events.RemoveOngoingEvent(s.ID, domain.EventKindAttack)
			// outcome unknown without snapshot — skip notify for now
		}
	}

	// current active events not in store → new
	storedIDs := make(map[int]struct{})
	for _, s := range stored {
		storedIDs[s.ID] = struct{}{}
	}

	for _, e := range current.AttackEvents {
		if e.Status != domain.EventStatusActive {
			continue
		}
		if _, exists := storedIDs[e.ID]; !exists {
			p.events.SaveOngoingEvent(e.ID, domain.EventKindAttack)
			attackCopy := e
			p.notify(domain.EventMessage{
				Kind:        domain.EventKindAttack,
				Transition:  domain.EventTransitionStarted,
				AttackEvent: &attackCopy,
			})
		}
	}
}
