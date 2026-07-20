package stdout

import (
	"fmt"

	"github.com/ametis70/hellbot/internal/domain"
)

func formatMessage(msg domain.EventMessage) (string, error) {
	var formattedMessage string

	switch msg.Kind {
	case domain.EventKindDefend:
		formattedMessage = fmt.Sprintf(
			"[defend] %s — Enemy: %v, Region: %d, Start: %v, End: %v",
			msg.Transition,
			msg.DefendEvent.Enemy,
			msg.DefendEvent.Region,
			msg.DefendEvent.StartTime,
			msg.DefendEvent.EndTime,
		)
	case domain.EventKindAttack:
		formattedMessage = fmt.Sprintf(
			"[attack] %s — Enemy: %v, Start: %v, End: %v",
			msg.Transition,
			msg.AttackEvent.Enemy,
			msg.AttackEvent.StartTime,
			msg.AttackEvent.EndTime,
		)
	}

	if formattedMessage == "" {
		return "", fmt.Errorf("unknown event kind: %s", msg.Kind)
	}

	return formattedMessage, nil
}

type StdoutNotifier struct{}

func New() *StdoutNotifier {
	return &StdoutNotifier{}
}

func (n *StdoutNotifier) Notify(msg domain.EventMessage) error {
	formattedMessage, err := formatMessage(msg)
	if err != nil {
		return err
	}

	fmt.Println(formattedMessage)
	return nil
}
