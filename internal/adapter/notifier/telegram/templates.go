package telegram

import (
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

// DefaultTemplates returns the default Telegram message templates.
// Times use {END_TIME_FORMATTED} rendered in the configured timezone.
func DefaultTemplates() domain.Templates {
	return domain.Templates{
		DefendRegionStarted:       "⚔️ *The {FACTION} is attacking {REGION_NAME} \\({REGION_NUMBER}/{TOTAL_REGIONS}\\)\\!*\nEnds: {END_TIME_FORMATTED}",
		DefendSuperEarthStarted:   "🚨 *The {FACTION} is attacking Super Earth\\!*\nEnds: {END_TIME_FORMATTED}",
		DefendRegionSucceeded:     "✅ *{REGION_NAME} \\({REGION_NUMBER}/{TOTAL_REGIONS}\\) has been defended against the {FACTION}\\!*",
		DefendSuperEarthSucceeded: "✅ *Super Earth has been defended against the {FACTION}\\!*",
		DefendRegionFailed:        "❌ *{REGION_NAME} \\({REGION_NUMBER}/{TOTAL_REGIONS}\\) has fallen to the {FACTION}\\.*",
		DefendSuperEarthFailed:    "❌ *Super Earth has fallen to the {FACTION}\\.*",
		AttackHomeworldStarted:    "🚀 *An attack against the {FACTION}'s homeworld has started\\!*\nEnds: {END_TIME_FORMATTED}",
		AttackSucceeded:           "✅ *Attack succeeded\\! The {FACTION} were defeated\\.*",
		AttackFailed:              "❌ *Attack failed\\! The {FACTION} defended their homeworld\\.*",
	}
}

// TimeFormatter returns a time formatter for Telegram using the given timezone.
func TimeFormatter(loc *time.Location) func(time.Time) string {
	return func(t time.Time) string {
		return escape(t.In(loc).Format("2006-01-02 15:04 MST"))
	}
}
