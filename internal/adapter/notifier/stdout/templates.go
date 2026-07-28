package stdout

import (
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

// DefaultTemplates returns the default stdout message templates.
// Times use RFC3339 format in the configured timezone.
func DefaultTemplates() domain.Templates {
	return domain.Templates{
		DefendRegionStarted:       "[defend] started — {FACTION} are attacking {REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS}), ends {END_TIME_FORMATTED}",
		DefendSuperEarthStarted:   "[defend] started — {FACTION} are attacking Super Earth, ends {END_TIME_FORMATTED}",
		DefendRegionSucceeded:     "[defend] succeeded — {REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS}) held against {FACTION}",
		DefendSuperEarthSucceeded: "[defend] succeeded — Super Earth held against {FACTION}",
		DefendRegionFailed:        "[defend] failed — {REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS}) fell to {FACTION}",
		DefendSuperEarthFailed:    "[defend] failed — Super Earth fell to {FACTION}",
		AttackHomeworldStarted:    "[attack] started — against {FACTION} homeworld, ends {END_TIME_FORMATTED}",
		AttackSucceeded:           "[attack] succeeded — {FACTION} defeated",
		AttackFailed:              "[attack] failed — {FACTION} defended homeworld",
	}
}

// TimeFormatter returns a time formatter for stdout using the given timezone.
func TimeFormatter(loc *time.Location) func(time.Time) string {
	return func(t time.Time) string {
		return formatTime(t, loc)
	}
}
