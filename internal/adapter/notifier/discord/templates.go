package discord

import (
	"fmt"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

// DefaultTemplates returns the default Discord message templates.
// Times use Discord's native <t:UNIX:f> format which renders in the viewer's local timezone.
func DefaultTemplates() domain.Templates {
	return domain.Templates{
		DefendRegionStarted:       "⚔️ **The {FACTION} are attacking {REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS})!**\nEnds: <t:{END_TIME_UNIX}:f>",
		DefendSuperEarthStarted:   "🚨 **The {FACTION} are attacking Super Earth!**\nEnds: <t:{END_TIME_UNIX}:f>",
		DefendRegionSucceeded:     "✅ **{REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS}) has been defended against the {FACTION}!**",
		DefendSuperEarthSucceeded: "✅ **Super Earth has been defended against the {FACTION}!**",
		DefendRegionFailed:        "❌ **{REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS}) has fallen to the {FACTION}.**",
		DefendSuperEarthFailed:    "❌ **Super Earth has fallen to the {FACTION}.**",
		AttackHomeworldStarted:    "🚀 **An attack against the {FACTION}'s homeworld has started!**\nEnds: <t:{END_TIME_UNIX}:f>",
		AttackSucceeded:           "✅ **Attack succeeded! The {FACTION} were defeated.**",
		AttackFailed:              "❌ **Attack failed! The {FACTION} defended their homeworld.**",
	}
}

// TimeFormatter returns a time formatter for Discord.
// Discord uses <t:UNIX:f> for native timestamp rendering,
// so the formatted time is only used when {END_TIME_FORMATTED} or
// {START_TIME_FORMATTED} are used in custom templates.
func TimeFormatter(loc *time.Location) func(time.Time) string {
	return func(t time.Time) string {
		return fmt.Sprintf("<t:%d:f>", t.Unix())
	}
}
