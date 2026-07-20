package port

import "github.com/ametis70/hellbot/internal/domain"

type Fetcher interface {
	FetchCampaign() (*domain.CampaignStatus, error)
}
