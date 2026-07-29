package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

// --- ParseEnemy ---

func TestParseEnemy(t *testing.T) {
	cases := []struct {
		input string
		want  domain.Enemy
		ok    bool
	}{
		{"bugs", domain.EnemyBug, true},
		{"bug", domain.EnemyBug, true},
		{"BUGS", domain.EnemyBug, true},
		{"cyborgs", domain.EnemyCyborg, true},
		{"cyborg", domain.EnemyCyborg, true},
		{"illuminate", domain.EnemyIlluminate, true},
		{"ILLUMINATE", domain.EnemyIlluminate, true},
		{"unknown", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := domain.ParseEnemy(c.input)
		if ok != c.ok {
			t.Errorf("ParseEnemy(%q): ok=%v want %v", c.input, ok, c.ok)
		}
		if ok && got != c.want {
			t.Errorf("ParseEnemy(%q): got %v want %v", c.input, got, c.want)
		}
	}
}

// --- BuildWarVars ---

func TestBuildWarVars(t *testing.T) {
	vars := domain.BuildWarVars(&domain.WarEvent{Season: 42})
	if vars.Season != "42" {
		t.Errorf("expected Season=42, got %q", vars.Season)
	}
}

// --- MergeTemplates ---

func TestMergeTemplates_OverridesNonEmpty(t *testing.T) {
	defaults := domain.Templates{
		DefendRegionStarted: "default defend",
		AttackSucceeded:     "default attack",
		WarWon:              "default war won",
	}
	user := domain.Templates{
		DefendRegionStarted: "custom defend",
	}
	result := domain.MergeTemplates(defaults, user)
	if result.DefendRegionStarted != "custom defend" {
		t.Errorf("expected custom defend, got %q", result.DefendRegionStarted)
	}
	if result.AttackSucceeded != "default attack" {
		t.Errorf("expected default attack to be preserved, got %q", result.AttackSucceeded)
	}
}

func TestMergeTemplates_AllFields(t *testing.T) {
	defaults := domain.Templates{}
	user := domain.Templates{
		DefendRegionStarted:       "a",
		DefendSuperEarthStarted:   "b",
		DefendRegionSucceeded:     "c",
		DefendSuperEarthSucceeded: "d",
		DefendRegionFailed:        "e",
		DefendSuperEarthFailed:    "f",
		AttackHomeworldStarted:    "g",
		AttackSucceeded:           "h",
		AttackFailed:              "i",
		WarWon:                    "j",
		WarLost:                   "k",
	}
	result := domain.MergeTemplates(defaults, user)
	if result.DefendRegionStarted != "a" || result.WarLost != "k" {
		t.Error("MergeTemplates: not all fields overridden")
	}
}

// --- RenderEvent ---

func timeFormatter(_ time.Time) string { return "T" }

func TestRenderEvent_AttackStarted(t *testing.T) {
	tmpl := domain.Templates{AttackHomeworldStarted: "attack {FACTION} started"}
	ev := domain.AttackEvent{Enemy: domain.EnemyCyborg, Status: domain.EventStatusActive}
	msg := domain.EventMessage{Kind: domain.EventKindAttack, Transition: domain.EventTransitionStarted, AttackEvent: &ev}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "attack Cyborgs started" {
		t.Errorf("unexpected render: %q", got)
	}
}

func TestRenderEvent_AttackSucceeded(t *testing.T) {
	tmpl := domain.Templates{AttackSucceeded: "won {FACTION}"}
	ev := domain.AttackEvent{Enemy: domain.EnemyBug}
	msg := domain.EventMessage{Kind: domain.EventKindAttack, Transition: domain.EventTransitionSucceeded, AttackEvent: &ev}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil || got != "won Bugs" {
		t.Errorf("unexpected: err=%v got=%q", err, got)
	}
}

func TestRenderEvent_AttackFailed(t *testing.T) {
	tmpl := domain.Templates{AttackFailed: "lost {FACTION}"}
	ev := domain.AttackEvent{Enemy: domain.EnemyIlluminate}
	msg := domain.EventMessage{Kind: domain.EventKindAttack, Transition: domain.EventTransitionFailed, AttackEvent: &ev}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil || got != "lost Illuminate" {
		t.Errorf("unexpected: err=%v got=%q", err, got)
	}
}

func TestRenderEvent_AttackNilEvent(t *testing.T) {
	tmpl := domain.Templates{}
	msg := domain.EventMessage{Kind: domain.EventKindAttack, Transition: domain.EventTransitionStarted}
	_, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err == nil {
		t.Error("expected error for nil attack event")
	}
}

func TestRenderEvent_DefendRegionStarted(t *testing.T) {
	tmpl := domain.Templates{DefendRegionStarted: "defend {FACTION} {REGION_NAME}"}
	ev := &domain.DefendEvent{Region: 1, Enemy: domain.EnemyCyborg}
	msg := domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionStarted, DefendEvent: ev}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "Cyborgs") {
		t.Errorf("expected Cyborgs in output, got %q", got)
	}
}

func TestRenderEvent_DefendSuperEarthStarted(t *testing.T) {
	tmpl := domain.Templates{DefendSuperEarthStarted: "super earth {FACTION}"}
	ev := &domain.DefendEvent{Region: 0, Enemy: domain.EnemyIlluminate}
	msg := domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionStarted, DefendEvent: ev}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil || got != "super earth Illuminate" {
		t.Errorf("unexpected: err=%v got=%q", err, got)
	}
}

func TestRenderEvent_DefendRegionSucceeded(t *testing.T) {
	tmpl := domain.Templates{DefendRegionSucceeded: "defended {FACTION}"}
	ev := &domain.DefendEvent{Region: 2, Enemy: domain.EnemyBug}
	msg := domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionSucceeded, DefendEvent: ev}
	_, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderEvent_DefendSuperEarthSucceeded(t *testing.T) {
	tmpl := domain.Templates{DefendSuperEarthSucceeded: "se defended"}
	ev := &domain.DefendEvent{Region: 0, Enemy: domain.EnemyBug}
	msg := domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionSucceeded, DefendEvent: ev}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil || got != "se defended" {
		t.Errorf("unexpected: err=%v got=%q", err, got)
	}
}

func TestRenderEvent_DefendRegionFailed(t *testing.T) {
	tmpl := domain.Templates{DefendRegionFailed: "lost region"}
	ev := &domain.DefendEvent{Region: 3, Enemy: domain.EnemyCyborg}
	msg := domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionFailed, DefendEvent: ev}
	_, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderEvent_DefendSuperEarthFailed(t *testing.T) {
	tmpl := domain.Templates{DefendSuperEarthFailed: "se lost"}
	ev := &domain.DefendEvent{Region: 0, Enemy: domain.EnemyCyborg}
	msg := domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionFailed, DefendEvent: ev}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil || got != "se lost" {
		t.Errorf("unexpected: err=%v got=%q", err, got)
	}
}

func TestRenderEvent_DefendNilEvent(t *testing.T) {
	msg := domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionStarted}
	_, err := domain.RenderEvent(domain.Templates{}, msg, timeFormatter)
	if err == nil {
		t.Error("expected error for nil defend event")
	}
}

func TestRenderEvent_WarWon(t *testing.T) {
	tmpl := domain.Templates{WarWon: "war won {SEASON}"}
	msg := domain.EventMessage{Kind: domain.EventKindWar, Transition: domain.EventTransitionSucceeded, WarEvent: &domain.WarEvent{Season: 99}}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil || got != "war won 99" {
		t.Errorf("unexpected: err=%v got=%q", err, got)
	}
}

func TestRenderEvent_WarLost(t *testing.T) {
	tmpl := domain.Templates{WarLost: "war lost {SEASON}"}
	msg := domain.EventMessage{Kind: domain.EventKindWar, Transition: domain.EventTransitionFailed, WarEvent: &domain.WarEvent{Season: 7}}
	got, err := domain.RenderEvent(tmpl, msg, timeFormatter)
	if err != nil || got != "war lost 7" {
		t.Errorf("unexpected: err=%v got=%q", err, got)
	}
}

func TestRenderEvent_WarNilEvent(t *testing.T) {
	msg := domain.EventMessage{Kind: domain.EventKindWar, Transition: domain.EventTransitionSucceeded}
	_, err := domain.RenderEvent(domain.Templates{}, msg, timeFormatter)
	if err == nil {
		t.Error("expected error for nil war event")
	}
}

func TestRenderEvent_UnknownKind(t *testing.T) {
	msg := domain.EventMessage{Kind: "unknown", Transition: "started"}
	_, err := domain.RenderEvent(domain.Templates{}, msg, timeFormatter)
	if err == nil {
		t.Error("expected error for unknown event kind")
	}
}

// --- FormatStatus ---

func TestFormatStatus_ContainsSeason(t *testing.T) {
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 42, Enemy: domain.EnemyBug, Status: domain.FactionStatusActive, Points: 100, PointsMax: 1000},
		},
	}
	out := domain.FormatStatus(c, nil)
	if !strings.Contains(out, "42") {
		t.Errorf("expected season 42 in output, got:\n%s", out)
	}
}

func TestFormatStatus_FilterByFaction(t *testing.T) {
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 1, Enemy: domain.EnemyBug, Status: domain.FactionStatusActive, Points: 100, PointsMax: 1000},
			{Season: 1, Enemy: domain.EnemyCyborg, Status: domain.FactionStatusActive, Points: 200, PointsMax: 1000},
		},
	}
	filter := domain.EnemyBug
	out := domain.FormatStatus(c, &filter)
	if strings.Contains(out, "Cyborgs") {
		t.Errorf("expected Cyborgs to be filtered out, got:\n%s", out)
	}
}

func TestFormatStatus_DefeatedFaction(t *testing.T) {
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 1, Enemy: domain.EnemyBug, Status: domain.FactionStatusDefeated},
		},
	}
	out := domain.FormatStatus(c, nil)
	if !strings.Contains(out, "defeated") {
		t.Errorf("expected 'defeated' in output, got:\n%s", out)
	}
}

func TestFormatStatus_HiddenFaction(t *testing.T) {
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 1, Enemy: domain.EnemyCyborg, Status: domain.FactionStatusHidden},
		},
	}
	out := domain.FormatStatus(c, nil)
	if !strings.Contains(out, "hidden") {
		t.Errorf("expected 'hidden' in output, got:\n%s", out)
	}
}

func TestFormatStatus_ActiveDefendEvent(t *testing.T) {
	endTime := time.Now().Add(2 * time.Hour)
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 1, Enemy: domain.EnemyBug, Status: domain.FactionStatusActive, Points: 0, PointsMax: 1000},
		},
		DefendEvent: &domain.DefendEvent{
			Region:  3,
			Enemy:   domain.EnemyBug,
			EndTime: endTime,
			Status:  domain.EventStatusActive,
		},
	}
	out := domain.FormatStatus(c, nil)
	if !strings.Contains(out, "defending") {
		t.Errorf("expected 'defending' annotation, got:\n%s", out)
	}
}

func TestFormatStatus_ActiveDefendSuperEarth(t *testing.T) {
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 1, Enemy: domain.EnemyBug, Status: domain.FactionStatusActive, Points: 0, PointsMax: 1000},
		},
		DefendEvent: &domain.DefendEvent{
			Region:  0,
			Enemy:   domain.EnemyBug,
			EndTime: time.Now().Add(time.Hour),
			Status:  domain.EventStatusActive,
		},
	}
	out := domain.FormatStatus(c, nil)
	if !strings.Contains(out, "Super Earth") {
		t.Errorf("expected Super Earth in output, got:\n%s", out)
	}
}

func TestFormatStatus_ActiveAttackEvent(t *testing.T) {
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 1, Enemy: domain.EnemyCyborg, Status: domain.FactionStatusActive, Points: 100, PointsMax: 1000},
		},
		AttackEvents: []domain.AttackEvent{
			{Enemy: domain.EnemyCyborg, Status: domain.EventStatusActive, EndTime: time.Now().Add(time.Hour)},
		},
	}
	out := domain.FormatStatus(c, nil)
	if !strings.Contains(out, "attacking") {
		t.Errorf("expected 'attacking' in output, got:\n%s", out)
	}
}

func TestFormatStatus_NoFactions(t *testing.T) {
	c := &domain.CampaignStatus{}
	out := domain.FormatStatus(c, nil)
	if !strings.Contains(out, "War 0") {
		t.Errorf("expected 'War 0', got:\n%s", out)
	}
}

// --- FormatStatistics ---

func TestFormatStatistics_NoStats(t *testing.T) {
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{{Season: 5}},
	}
	out := domain.FormatStatistics(c)
	if !strings.Contains(out, "No statistics") {
		t.Errorf("expected 'No statistics', got:\n%s", out)
	}
}

func TestFormatStatistics_WithStats(t *testing.T) {
	c := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{{Season: 10}},
		Statistics: []domain.Statistics{
			{
				Season: 10, Players: 5000, TotalUniquePlayers: 100000,
				Kills: 1000000, Deaths: 50000, Accidentals: 1000,
				Shots: 10000000, Hits: 5000000,
				Missions: 100, SuccessfulMissions: 80,
				DefendEvents: 10, SuccessfulDefendEvents: 7,
				AttackEvents: 5, SuccessfulAttackEvents: 3,
				CompletedPlanets: 20,
			},
		},
	}
	out := domain.FormatStatistics(c)
	if !strings.Contains(out, "War 10") {
		t.Errorf("expected 'War 10', got:\n%s", out)
	}
	if !strings.Contains(out, "Kills") {
		t.Errorf("expected 'Kills' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "50%") {
		t.Errorf("expected 50%% accuracy, got:\n%s", out)
	}
}

func TestFormatStatistics_ZeroDenominator(t *testing.T) {
	c := &domain.CampaignStatus{
		Statistics: []domain.Statistics{{Shots: 0, Hits: 0}},
	}
	// Should not panic on zero denominators.
	out := domain.FormatStatistics(c)
	if !strings.Contains(out, "0%") {
		t.Errorf("expected 0%% accuracy, got:\n%s", out)
	}
}

// --- IsHomeworld ---

func TestIsHomeworld(t *testing.T) {
	if !domain.IsHomeworld(11) {
		t.Error("expected region 11 to be homeworld")
	}
	if domain.IsHomeworld(0) {
		t.Error("expected region 0 to not be homeworld")
	}
}

// --- GetRegion fallback ---

func TestGetRegion_Fallback(t *testing.T) {
	r := domain.GetRegion(domain.EnemyBug, 999)
	if r.Name != "Region 999" {
		t.Errorf("expected fallback name, got %q", r.Name)
	}
}

// --- Enemy.String unknown ---

func TestEnemyString_Unknown(t *testing.T) {
	s := domain.Enemy(99).String()
	if !strings.Contains(s, "Unknown") {
		t.Errorf("expected Unknown in string, got %q", s)
	}
}
