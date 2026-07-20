package port

import "github.com/ametis70/hellbot/internal/domain"

type Notifier interface {
	Notify(msg domain.EventMessage) error
}
