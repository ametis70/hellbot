package domain

import (
	"fmt"
	"strings"
	"time"
)

// ParseEnemy converts a faction name string (case-insensitive) to an Enemy value.
// Returns (enemy, true) on match, (0, false) if not recognised.
func ParseEnemy(s string) (Enemy, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "bugs", "bug":
		return EnemyBug, true
	case "cyborgs", "cyborg":
		return EnemyCyborg, true
	case "illuminate":
		return EnemyIlluminate, true
	default:
		return 0, false
	}
}

// FormatStatus returns a human-readable war status string.
// If filter is non-nil, only the matching faction is shown.
func FormatStatus(c *CampaignStatus, filter *Enemy) string {
	var sb strings.Builder

	// Determine season from first faction or fall back.
	season := 0
	if len(c.FactionsStatus) > 0 {
		season = c.FactionsStatus[0].Season
	}
	fmt.Fprintf(&sb, "War %d — Status\n\n", season)

	// Per-faction progress.
	for _, f := range c.FactionsStatus {
		if filter != nil && f.Enemy != *filter {
			continue
		}
		sb.WriteString(formatFactionStatus(f, c))
	}

	// Active events.
	var events []string
	if c.DefendEvent != nil && c.DefendEvent.Status == EventStatusActive {
		e := c.DefendEvent
		if filter == nil || e.Enemy == *filter {
			if IsSuperEarth(e.Region) {
				events = append(events, fmt.Sprintf(
					"⚔️  The %s are attacking Super Earth — ends %s",
					e.Enemy, formatRelativeTime(e.EndTime),
				))
			} else {
				region := GetRegion(e.Enemy, e.Region)
				events = append(events, fmt.Sprintf(
					"⚔️  The %s are attacking sector %d: %s — ends %s",
					e.Enemy, e.Region, region.Name, formatRelativeTime(e.EndTime),
				))
			}
		}
	}
	for _, e := range c.AttackEvents {
		if e.Status != EventStatusActive {
			continue
		}
		if filter != nil && e.Enemy != *filter {
			continue
		}
		events = append(events, fmt.Sprintf(
			"🚀 Attacking %s homeworld — ends %s",
			e.Enemy, formatRelativeTime(e.EndTime),
		))
	}

	if len(events) > 0 {
		sb.WriteString("\nActive events:\n")
		for _, ev := range events {
			sb.WriteString("  " + ev + "\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func formatFactionStatus(f FactionStatus, c *CampaignStatus) string {
	switch f.Status {
	case FactionStatusDefeated:
		return fmt.Sprintf("%-16s defeated\n", "The "+f.Enemy.String())
	case FactionStatusHidden:
		return fmt.Sprintf("%-16s hidden\n", "The "+f.Enemy.String())
	}

	// Overall war progress: points counts down from pointsMax as helldivers advance.
	totalPct := 0
	if f.PointsMax > 0 {
		totalPct = f.Points * 100 / f.PointsMax
		if totalPct > 100 {
			totalPct = 100
		}
	}
	totalBar := progressBar(totalPct, 10)

	// Current sector: sectorsEarned = floor(points / pointsPerSector)
	// displayed as sectorsEarned/TotalRegions (e.g. "6/10").
	// sectorPoints = points - sectorsEarned * pointsPerSector (remainder within current sector).
	sectorNum := 0
	sectorPct := 0
	sectorPoints := 0
	sectorPointsMax := 0
	if f.PointsMax > 0 {
		pointsPerSector := f.PointsMax / TotalRegions
		if pointsPerSector > 0 {
			sectorsEarned := f.Points / pointsPerSector
			sectorNum = sectorsEarned
			if sectorNum > TotalRegions {
				sectorNum = TotalRegions
			}
			sectorPointsMax = pointsPerSector
			sectorPoints = f.Points - sectorsEarned*pointsPerSector
			sectorPct = sectorPoints * 100 / pointsPerSector
		}
	}

	region := GetRegion(f.Enemy, sectorNum+1)

	// Event annotation for the current sector.
	eventNote := ""
	if c.DefendEvent != nil && c.DefendEvent.Enemy == f.Enemy && c.DefendEvent.Status == EventStatusActive {
		e := c.DefendEvent
		if IsSuperEarth(e.Region) {
			eventNote = " ⚔️ defending Super Earth"
		} else {
			defRegion := GetRegion(e.Enemy, e.Region)
			eventNote = fmt.Sprintf(" ⚔️ defending %s", defRegion.Name)
		}
	}
	for _, e := range c.AttackEvents {
		if e.Enemy == f.Enemy && e.Status == EventStatusActive {
			eventNote = " 🚀 attacking homeworld"
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "The %s (active)\n", f.Enemy.String())
	fmt.Fprintf(&sb, "  %s %3d%%\n", totalBar, totalPct)
	fmt.Fprintf(&sb, "  %s / %s pts\n", fmtInt(f.Points), fmtInt(f.PointsMax))
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "  Sector %d/11: %s%s\n", sectorNum+1, region.Name, eventNote)
	fmt.Fprintf(&sb, "  %s %3d%%\n", progressBar(sectorPct, 10), sectorPct)
	fmt.Fprintf(&sb, "  %s / %s pts\n", fmtInt(sectorPoints), fmtInt(sectorPointsMax))
	sb.WriteString("\n")
	return sb.String()
}

// activeSector is no longer used — sector logic is inlined in formatFactionStatus.

func progressBar(pct, width int) string {
	filled := pct * width / 100
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func formatRelativeTime(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "ended"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// FormatStatistics returns a human-readable statistics string with all factions summed.
func FormatStatistics(c *CampaignStatus) string {
	var sb strings.Builder

	season := 0
	if len(c.FactionsStatus) > 0 {
		season = c.FactionsStatus[0].Season
	}
	fmt.Fprintf(&sb, "War %d — Statistics\n\n", season)

	if len(c.Statistics) == 0 {
		sb.WriteString("No statistics available.\n")
		return strings.TrimRight(sb.String(), "\n")
	}

	var (
		players            int
		totalUniquePlayers int
		missions           int
		successfulMissions int
		completedPlanets   int
		defendEvents       int
		successfulDefend   int
		attackEvents       int
		successfulAttack   int
		deaths             int
		kills              int
		accidentals        int
		shots              int
		hits               int
	)

	for _, s := range c.Statistics {
		players += s.Players
		totalUniquePlayers += s.TotalUniquePlayers
		missions += s.Missions
		successfulMissions += s.SuccessfulMissions
		completedPlanets += s.CompletedPlanets
		defendEvents += s.DefendEvents
		successfulDefend += s.SuccessfulDefendEvents
		attackEvents += s.AttackEvents
		successfulAttack += s.SuccessfulAttackEvents
		deaths += s.Deaths
		kills += s.Kills
		accidentals += s.Accidentals
		shots += s.Shots
		hits += s.Hits
	}

	missionPct := pct(successfulMissions, missions)
	defendPct := pct(successfulDefend, defendEvents)
	attackPct := pct(successfulAttack, attackEvents)
	hitPct := pct(hits, shots)

	fmt.Fprintf(&sb, "Players online:     %s\n", fmtInt(players))
	fmt.Fprintf(&sb, "Total players:      %s\n", fmtInt(totalUniquePlayers))
	fmt.Fprintf(&sb, "Kills:              %s\n", fmtInt(kills))
	fmt.Fprintf(&sb, "Deaths:             %s\n", fmtInt(deaths))
	fmt.Fprintf(&sb, "Accidentals:        %s\n", fmtInt(accidentals))
	fmt.Fprintf(&sb, "Shots fired:        %s\n", fmtInt(shots))
	fmt.Fprintf(&sb, "Accuracy:           %d%%\n", hitPct)
	fmt.Fprintf(&sb, "Missions:           %s (%s successful, %d%%)\n", fmtInt(missions), fmtInt(successfulMissions), missionPct)
	fmt.Fprintf(&sb, "Defend events:      %s (%s successful, %d%%)\n", fmtInt(defendEvents), fmtInt(successfulDefend), defendPct)
	fmt.Fprintf(&sb, "Attack events:      %s (%s successful, %d%%)\n", fmtInt(attackEvents), fmtInt(successfulAttack), attackPct)
	fmt.Fprintf(&sb, "Planets liberated:  %s\n", fmtInt(completedPlanets))

	return strings.TrimRight(sb.String(), "\n")
}

func pct(num, denom int) int {
	if denom == 0 {
		return 0
	}
	return num * 100 / denom
}

func fmtInt(n int) string {
	s := fmt.Sprintf("%d", n)
	// Insert thousands separators.
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		pos := len(s) - i
		if i > 0 && pos%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}
