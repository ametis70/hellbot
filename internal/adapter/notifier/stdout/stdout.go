package stdout

import (
	"fmt"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

func formatTime(t time.Time, loc *time.Location) string {
	return t.In(loc).Format(time.RFC3339)
}

func formatMessage(msg domain.EventMessage, loc *time.Location) (string, error) {
	var formattedMessage string

	switch msg.Kind {
	case domain.EventKindDefend:
		formattedMessage = fmt.Sprintf(
			"[defend] %s — Enemy: %v, Region: %d, Start: %v, End: %v",
			msg.Transition,
			msg.DefendEvent.Enemy,
			msg.DefendEvent.Region,
			formatTime(msg.DefendEvent.StartTime, loc),
			formatTime(msg.DefendEvent.EndTime, loc),
		)
	case domain.EventKindAttack:
		formattedMessage = fmt.Sprintf(
			"[attack] %s — Enemy: %v, Start: %v, End: %v",
			msg.Transition,
			msg.AttackEvent.Enemy,
			formatTime(msg.AttackEvent.StartTime, loc),
			formatTime(msg.AttackEvent.EndTime, loc),
		)
	}

	if formattedMessage == "" {
		return "", fmt.Errorf("unknown event kind: %s", msg.Kind)
	}

	return formattedMessage, nil
}

type Options struct {
	Timezone *time.Location
}

type StdoutNotifier struct {
	opts Options
}

func New(opts Options) *StdoutNotifier {
	if opts.Timezone == nil {
		opts.Timezone = time.UTC
	}

	return &StdoutNotifier{opts: opts}
}

func (n *StdoutNotifier) Notify(msg domain.EventMessage) error {
	formattedMessage, err := formatMessage(msg, n.opts.Timezone)
	if err != nil {
		return err
	}

	fmt.Println(formattedMessage)
	return nil
}
