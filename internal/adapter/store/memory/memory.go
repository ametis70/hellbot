package memory

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ametis70/hellbot/internal/domain"
)

type MemoryStore struct {
	mu       sync.RWMutex
	campaign *domain.CampaignStatus
	events   map[string]*domain.OngoingEvent
}

func New() *MemoryStore {
	return &MemoryStore{
		events: make(map[string]*domain.OngoingEvent, 4),
	}
}

func eventKey(id int, kind domain.EventKind) string {
	return fmt.Sprintf("%d:%s", id, kind)
}

func (s *MemoryStore) SaveCampaign(c *domain.CampaignStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.campaign = c
	return nil
}

func (s *MemoryStore) LatestCampaign() (*domain.CampaignStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.campaign == nil {
		return nil, errors.New("no campaign stored")
	}
	return s.campaign, nil
}

func (s *MemoryStore) SaveOngoingEvent(id int, kind domain.EventKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[eventKey(id, kind)] = &domain.OngoingEvent{ID: id, Kind: kind}
	return nil
}

func (s *MemoryStore) RemoveOngoingEvent(id int, kind domain.EventKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[eventKey(id, kind)]; !ok {
		return errors.New("event not found")
	}

	delete(s.events, eventKey(id, kind))
	return nil
}

func (s *MemoryStore) GetOngoingEvent(id int, kind domain.EventKind) (*domain.OngoingEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	event, ok := s.events[eventKey(id, kind)]
	if !ok {
		return nil, errors.New("event not found")
	}
	return event, nil
}
