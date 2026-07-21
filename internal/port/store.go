package port

import "github.com/ametis70/hellbot/internal/domain"

type CampaignStore interface {
	SaveCampaign(c *domain.CampaignStatus) error
	LatestCampaign() (*domain.CampaignStatus, error)
}

type EventStore interface {
	SaveOngoingEvent(id int, kind domain.EventKind) error
	RemoveOngoingEvent(id int, kind domain.EventKind) error
	GetOngoingEvent(id int, kind domain.EventKind) (*domain.OngoingEvent, error)
	ListOngoingEvents(kind domain.EventKind) ([]*domain.OngoingEvent, error)
}
