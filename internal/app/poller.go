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
		changed := p.handleEvents(current, previous)
		if !changed {
			p.logger.Info("no changes since last fetch")
		}
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

func (p *Poller) handleEvents(current, previous *domain.CampaignStatus) bool {
	defendEventsChanged := p.handleDefendEvent(current, previous)
	attackEventsChanged := p.handleAttackEvents(current)
	warEventsChanged := p.handleWarEvents(current, previous)

	return defendEventsChanged || attackEventsChanged || warEventsChanged
}

func (p *Poller) handleDefendEvent(current, previous *domain.CampaignStatus) bool {
	stored, err := p.events.ListOngoingEvents(domain.EventKindDefend)
	if err != nil {
		p.logger.Error("failed to list ongoing defend events", "error", err)
		return false
	}

	// at most one defend event at a time
	var storedEvent *domain.OngoingEvent
	if len(stored) > 0 {
		storedEvent = stored[0]
	}

	if current.DefendEvent == nil && storedEvent == nil {
		return false
	}

	// no stored event yet — new event
	if current.DefendEvent != nil && storedEvent == nil {
		if current.DefendEvent.Status != domain.EventStatusActive {
			return false
		}
		if err := p.events.SaveOngoingEvent(current.DefendEvent.ID, domain.EventKindDefend); err != nil {
			p.logger.Error("failed to save ongoing defend event", "error", err)
			return false
		}
		p.notify(domain.EventMessage{
			Kind:        domain.EventKindDefend,
			Transition:  domain.EventTransitionStarted,
			DefendEvent: current.DefendEvent,
		})
		return true
	}

	// same event — check if status changed from active to ended
	if current.DefendEvent != nil && current.DefendEvent.ID == storedEvent.ID {
		if current.DefendEvent.Status == domain.EventStatusActive {
			return false
		}
		if err := p.events.RemoveOngoingEvent(storedEvent.ID, domain.EventKindDefend); err != nil {
			p.logger.Error("failed to remove ongoing defend event", "error", err)
			return false
		}
		transition := domain.EventTransitionFailed
		if current.DefendEvent.Status == domain.EventStatusSuccess {
			transition = domain.EventTransitionSucceeded
		}
		p.notify(domain.EventMessage{
			Kind:        domain.EventKindDefend,
			Transition:  transition,
			DefendEvent: current.DefendEvent,
		})
		return true
	}

	// different event ID — old ended, new started
	if current.DefendEvent != nil && current.DefendEvent.ID != storedEvent.ID {
		if previous.DefendEvent != nil && previous.DefendEvent.ID == storedEvent.ID {
			transition := domain.EventTransitionFailed
			if previous.DefendEvent.Status == domain.EventStatusSuccess {
				transition = domain.EventTransitionSucceeded
			}
			p.notify(domain.EventMessage{
				Kind:        domain.EventKindDefend,
				Transition:  transition,
				DefendEvent: previous.DefendEvent,
			})
		}
		if err := p.events.RemoveOngoingEvent(storedEvent.ID, domain.EventKindDefend); err != nil {
			p.logger.Error("failed to remove ongoing defend event", "error", err)
			return false
		}
		if current.DefendEvent.Status == domain.EventStatusActive {
			if err := p.events.SaveOngoingEvent(current.DefendEvent.ID, domain.EventKindDefend); err != nil {
				p.logger.Error("failed to save ongoing defend event", "error", err)
				return false
			}
			p.notify(domain.EventMessage{
				Kind:        domain.EventKindDefend,
				Transition:  domain.EventTransitionStarted,
				DefendEvent: current.DefendEvent,
			})
		}
		return true
	}

	return false
}

func (p *Poller) handleAttackEvents(current *domain.CampaignStatus) bool {
	stored, err := p.events.ListOngoingEvents(domain.EventKindAttack)
	if err != nil {
		p.logger.Error("failed to list ongoing attack events", "error", err)
		return false
	}

	// build map of current active attack IDs for O(1) lookup
	currentActive := make(map[int]domain.AttackEvent)
	for _, e := range current.AttackEvents {
		if e.Status == domain.EventStatusActive {
			currentActive[e.ID] = e
		}
	}

	changed := false

	// stored events not in current active → ended
	for _, s := range stored {
		if _, stillActive := currentActive[s.ID]; !stillActive {
			if err := p.events.RemoveOngoingEvent(s.ID, domain.EventKindAttack); err != nil {
				p.logger.Error("failed to remove ongoing attack event", "error", err)
				continue
			}

			for _, e := range current.AttackEvents {
				if e.ID == s.ID {
					transition := domain.EventTransitionFailed
					if e.Status == domain.EventStatusSuccess {
						transition = domain.EventTransitionSucceeded
					}
					attackCopy := e
					p.notify(domain.EventMessage{
						Kind:        domain.EventKindAttack,
						Transition:  transition,
						AttackEvent: &attackCopy,
					})
					break
				}
			}
			changed = true
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
			if err := p.events.SaveOngoingEvent(e.ID, domain.EventKindAttack); err != nil {
				p.logger.Error("failed to save ongoing attack event", "error", err)
				continue
			}
			attackCopy := e
			p.notify(domain.EventMessage{
				Kind:        domain.EventKindAttack,
				Transition:  domain.EventTransitionStarted,
				AttackEvent: &attackCopy,
			})
			changed = true
		}
	}
	return changed
}

func (p *Poller) handleWarEvents(current, previous *domain.CampaignStatus) bool {
	if len(previous.FactionsStatus) == 0 || len(current.FactionsStatus) == 0 {
		return false
	}

	prevSeason := previous.FactionsStatus[0].Season
	currSeason := current.FactionsStatus[0].Season
	if prevSeason == currSeason {
		return false
	}

	// Season changed — determine outcome from the previous season's final state.
	// War is won if all non-hidden factions were defeated.
	allDefeated := true
	for _, f := range previous.FactionsStatus {
		if f.Status == domain.FactionStatusHidden {
			continue
		}
		if f.Status != domain.FactionStatusDefeated {
			allDefeated = false
			break
		}
	}

	transition := domain.EventTransitionFailed
	if allDefeated {
		transition = domain.EventTransitionSucceeded
	}

	p.notify(domain.EventMessage{
		Kind:       domain.EventKindWar,
		Transition: transition,
		WarEvent:   &domain.WarEvent{Season: prevSeason},
	})
	return true
}
