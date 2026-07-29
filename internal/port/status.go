package port

import "github.com/ametis70/hellbot/internal/domain"

// StatusProvider gives read access to the latest cached campaign state.
type StatusProvider interface {
	LatestCampaign() (*domain.CampaignStatus, error)
}
