package domain

import (
	"fmt"
	"strings"
	"time"
)

// Templates holds message templates for each event kind and transition.
// Each field is a string with {VARIABLE} placeholders.
// Empty fields fall back to adapter-specific defaults.
type Templates struct {
	DefendRegionStarted       string `yaml:"defend_region_started"`
	DefendSuperEarthStarted   string `yaml:"defend_super_earth_started"`
	DefendRegionSucceeded     string `yaml:"defend_region_succeeded"`
	DefendSuperEarthSucceeded string `yaml:"defend_super_earth_succeeded"`
	DefendRegionFailed        string `yaml:"defend_region_failed"`
	DefendSuperEarthFailed    string `yaml:"defend_super_earth_failed"`
	AttackHomeworldStarted    string `yaml:"attack_homeworld_started"`
	AttackSucceeded           string `yaml:"attack_succeeded"`
	AttackFailed              string `yaml:"attack_failed"`
	WarWon                    string `yaml:"war_won"`
	WarLost                   string `yaml:"war_lost"`
}

// MergeTemplates merges user-provided templates over defaults.
// Any non-empty field in user overrides the corresponding default field.
func MergeTemplates(defaults, user Templates) Templates {
	result := defaults
	if user.DefendRegionStarted != "" {
		result.DefendRegionStarted = user.DefendRegionStarted
	}
	if user.DefendSuperEarthStarted != "" {
		result.DefendSuperEarthStarted = user.DefendSuperEarthStarted
	}
	if user.DefendRegionSucceeded != "" {
		result.DefendRegionSucceeded = user.DefendRegionSucceeded
	}
	if user.DefendSuperEarthSucceeded != "" {
		result.DefendSuperEarthSucceeded = user.DefendSuperEarthSucceeded
	}
	if user.DefendRegionFailed != "" {
		result.DefendRegionFailed = user.DefendRegionFailed
	}
	if user.DefendSuperEarthFailed != "" {
		result.DefendSuperEarthFailed = user.DefendSuperEarthFailed
	}
	if user.AttackHomeworldStarted != "" {
		result.AttackHomeworldStarted = user.AttackHomeworldStarted
	}
	if user.AttackSucceeded != "" {
		result.AttackSucceeded = user.AttackSucceeded
	}
	if user.AttackFailed != "" {
		result.AttackFailed = user.AttackFailed
	}
	if user.WarWon != "" {
		result.WarWon = user.WarWon
	}
	if user.WarLost != "" {
		result.WarLost = user.WarLost
	}
	return result
}

// TemplateVars holds all substitution variables for template rendering.
type TemplateVars struct {
	Faction            string
	Season             string
	RegionName         string
	RegionNumber       string
	TotalRegions       string
	StartTimeFormatted string
	EndTimeFormatted   string
	StartTimeUnix      string
	EndTimeUnix        string
	Players            string
}

// Render substitutes all {VARIABLE} placeholders in a template string.
func Render(tmpl string, vars TemplateVars) string {
	r := strings.NewReplacer(
		"{FACTION}", vars.Faction,
		"{SEASON}", vars.Season,
		"{REGION_NAME}", vars.RegionName,
		"{REGION_NUMBER}", vars.RegionNumber,
		"{TOTAL_REGIONS}", vars.TotalRegions,
		"{START_TIME_FORMATTED}", vars.StartTimeFormatted,
		"{END_TIME_FORMATTED}", vars.EndTimeFormatted,
		"{START_TIME_UNIX}", vars.StartTimeUnix,
		"{END_TIME_UNIX}", vars.EndTimeUnix,
		"{PLAYERS}", vars.Players,
	)
	return r.Replace(tmpl)
}

// BuildDefendVars builds template variables for a defend event.
func BuildDefendVars(e *DefendEvent, formatTime func(time.Time) string) TemplateVars {
	region := GetRegion(e.Enemy, e.Region)
	return TemplateVars{
		Faction:            e.Enemy.String(),
		RegionName:         region.Name,
		RegionNumber:       fmt.Sprintf("%d", e.Region),
		TotalRegions:       fmt.Sprintf("%d", TotalRegions),
		StartTimeFormatted: formatTime(e.StartTime),
		EndTimeFormatted:   formatTime(e.EndTime),
		StartTimeUnix:      fmt.Sprintf("%d", e.StartTime.Unix()),
		EndTimeUnix:        fmt.Sprintf("%d", e.EndTime.Unix()),
		Players:            fmt.Sprintf("%d", e.PlayersAtStart),
	}
}

// BuildAttackVars builds template variables for an attack event.
func BuildAttackVars(e *AttackEvent, formatTime func(time.Time) string) TemplateVars {
	return TemplateVars{
		Faction:            e.Enemy.String(),
		StartTimeFormatted: formatTime(e.StartTime),
		EndTimeFormatted:   formatTime(e.EndTime),
		StartTimeUnix:      fmt.Sprintf("%d", e.StartTime.Unix()),
		EndTimeUnix:        fmt.Sprintf("%d", e.EndTime.Unix()),
		Players:            fmt.Sprintf("%d", e.PlayersAtStart),
	}
}

// BuildWarVars builds template variables for a war event.
func BuildWarVars(e *WarEvent) TemplateVars {
	return TemplateVars{
		Season: fmt.Sprintf("%d", e.Season),
	}
}

// RenderEvent picks the right template, builds vars, and renders the message.
func RenderEvent(templates Templates, msg EventMessage, formatTime func(time.Time) string) (string, error) {
	switch msg.Kind {
	case EventKindDefend:
		if msg.DefendEvent == nil {
			return "", fmt.Errorf("defend event is nil")
		}
		vars := BuildDefendVars(msg.DefendEvent, formatTime)
		switch msg.Transition {
		case EventTransitionStarted:
			if IsSuperEarth(msg.DefendEvent.Region) {
				return Render(templates.DefendSuperEarthStarted, vars), nil
			}
			return Render(templates.DefendRegionStarted, vars), nil
		case EventTransitionSucceeded:
			if IsSuperEarth(msg.DefendEvent.Region) {
				return Render(templates.DefendSuperEarthSucceeded, vars), nil
			}
			return Render(templates.DefendRegionSucceeded, vars), nil
		case EventTransitionFailed:
			if IsSuperEarth(msg.DefendEvent.Region) {
				return Render(templates.DefendSuperEarthFailed, vars), nil
			}
			return Render(templates.DefendRegionFailed, vars), nil
		}

	case EventKindAttack:
		if msg.AttackEvent == nil {
			return "", fmt.Errorf("attack event is nil")
		}
		vars := BuildAttackVars(msg.AttackEvent, formatTime)
		switch msg.Transition {
		case EventTransitionStarted:
			return Render(templates.AttackHomeworldStarted, vars), nil
		case EventTransitionSucceeded:
			return Render(templates.AttackSucceeded, vars), nil
		case EventTransitionFailed:
			return Render(templates.AttackFailed, vars), nil
		}

	case EventKindWar:
		if msg.WarEvent == nil {
			return "", fmt.Errorf("war event is nil")
		}
		vars := BuildWarVars(msg.WarEvent)
		switch msg.Transition {
		case EventTransitionSucceeded:
			return Render(templates.WarWon, vars), nil
		case EventTransitionFailed:
			return Render(templates.WarLost, vars), nil
		}
	}

	return "", fmt.Errorf("unhandled event kind=%s transition=%s", msg.Kind, msg.Transition)
}
