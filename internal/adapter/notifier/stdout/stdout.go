package stdout

import (
	"fmt"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

func formatTime(t time.Time, loc *time.Location) string {
	return t.In(loc).Format(time.RFC3339)
}

// Options holds configuration for the stdout notifier.
type Options struct {
	Timezone  *time.Location
	Templates *domain.Templates
}

// StdoutNotifier implements port.Notifier by printing to stdout.
type StdoutNotifier struct {
	opts      Options
	templates domain.Templates
}

// New creates a new StdoutNotifier.
func New(opts Options) *StdoutNotifier {
	if opts.Timezone == nil {
		opts.Timezone = time.UTC
	}

	templates := DefaultTemplates()
	if opts.Templates != nil {
		templates = domain.MergeTemplates(templates, *opts.Templates)
	}

	return &StdoutNotifier{
		opts:      opts,
		templates: templates,
	}
}

// Notify prints a formatted event message to stdout.
func (n *StdoutNotifier) Notify(msg domain.EventMessage) error {
	text, err := domain.RenderEvent(n.templates, msg, TimeFormatter(n.opts.Timezone))
	if err != nil {
		return err
	}
	fmt.Println(text)
	return nil
}

// formatMessage is kept for tests — delegates to domain.RenderEvent with default templates.
func formatMessage(msg domain.EventMessage, loc *time.Location) (string, error) {
	return domain.RenderEvent(DefaultTemplates(), msg, TimeFormatter(loc))
}
