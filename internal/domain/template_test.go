package domain

import (
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Unix(1784501941, 0).UTC()

func identityFormatter(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func fixedDefendEvent(region int, enemy Enemy) *DefendEvent {
	return &DefendEvent{
		Season:         159,
		ID:             5080,
		StartTime:      fixedTime,
		EndTime:        fixedTime.Add(48 * time.Hour),
		Region:         region,
		Enemy:          enemy,
		PointsMax:      31602,
		Points:         486,
		Status:         EventStatusActive,
		PlayersAtStart: 100,
	}
}

func fixedAttackEvent(enemy Enemy) *AttackEvent {
	return &AttackEvent{
		Season:         159,
		ID:             924,
		StartTime:      fixedTime,
		EndTime:        fixedTime.Add(48 * time.Hour),
		Enemy:          enemy,
		PointsMax:      31576,
		Points:         0,
		Status:         EventStatusActive,
		PlayersAtStart: 184,
	}
}

// --- Render tests ---

func TestRender_AllVars(t *testing.T) {
	vars := TemplateVars{
		Faction:            "Illuminate",
		RegionName:         "Orionis Region",
		RegionNumber:       "5",
		TotalRegions:       "10",
		StartTimeFormatted: "2026-07-19T19:59:01Z",
		EndTimeFormatted:   "2026-07-21T19:59:01Z",
		StartTimeUnix:      "1784501941",
		EndTimeUnix:        "1784674741",
		Players:            "100",
	}
	tmpl := "{FACTION} {REGION_NAME} {REGION_NUMBER}/{TOTAL_REGIONS} {START_TIME_FORMATTED} {END_TIME_FORMATTED} {START_TIME_UNIX} {END_TIME_UNIX} {PLAYERS}"
	result := Render(tmpl, vars)

	expected := "Illuminate Orionis Region 5/10 2026-07-19T19:59:01Z 2026-07-21T19:59:01Z 1784501941 1784674741 100"
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestRender_UnknownPlaceholderUnchanged(t *testing.T) {
	result := Render("hello {UNKNOWN_VAR}", TemplateVars{})
	if result != "hello {UNKNOWN_VAR}" {
		t.Errorf("expected unknown placeholder to be unchanged, got: %s", result)
	}
}

// --- MergeTemplates tests ---

func TestMergeTemplates_UserOverridesDefaults(t *testing.T) {
	defaults := Templates{DefendRegionStarted: "default"}
	user := Templates{DefendRegionStarted: "custom"}
	result := MergeTemplates(defaults, user)
	if result.DefendRegionStarted != "custom" {
		t.Errorf("expected 'custom', got %s", result.DefendRegionStarted)
	}
}

func TestMergeTemplates_EmptyUserKeepsDefault(t *testing.T) {
	defaults := Templates{DefendRegionStarted: "default"}
	user := Templates{}
	result := MergeTemplates(defaults, user)
	if result.DefendRegionStarted != "default" {
		t.Errorf("expected 'default', got %s", result.DefendRegionStarted)
	}
}

func TestMergeTemplates_PartialOverride(t *testing.T) {
	defaults := Templates{
		DefendRegionStarted:   "default-region",
		DefendRegionSucceeded: "default-succeeded",
	}
	user := Templates{DefendRegionStarted: "custom-region"}
	result := MergeTemplates(defaults, user)
	if result.DefendRegionStarted != "custom-region" {
		t.Errorf("expected 'custom-region', got %s", result.DefendRegionStarted)
	}
	if result.DefendRegionSucceeded != "default-succeeded" {
		t.Errorf("expected 'default-succeeded', got %s", result.DefendRegionSucceeded)
	}
}

// --- RenderEvent tests ---

func TestRenderEvent_DefendRegionStarted(t *testing.T) {
	templates := Templates{
		DefendRegionStarted: "{FACTION} attacking {REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS}) ends {END_TIME_FORMATTED}",
	}
	msg := EventMessage{
		Kind:        EventKindDefend,
		Transition:  EventTransitionStarted,
		DefendEvent: fixedDefendEvent(5, EnemyIlluminate),
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Illuminate") {
		t.Errorf("expected faction name, got: %s", result)
	}
	if !strings.Contains(result, "Orionis Region") {
		t.Errorf("expected region name, got: %s", result)
	}
	if !strings.Contains(result, "5/10") {
		t.Errorf("expected region position, got: %s", result)
	}
}

func TestRenderEvent_DefendSuperEarthStarted(t *testing.T) {
	templates := Templates{
		DefendSuperEarthStarted: "{FACTION} attacking Super Earth ends {END_TIME_FORMATTED}",
	}
	msg := EventMessage{
		Kind:        EventKindDefend,
		Transition:  EventTransitionStarted,
		DefendEvent: fixedDefendEvent(SuperEarthRegion, EnemyIlluminate),
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Super Earth") {
		t.Errorf("expected 'Super Earth', got: %s", result)
	}
}

func TestRenderEvent_DefendSucceeded(t *testing.T) {
	templates := Templates{DefendRegionSucceeded: "{REGION_NAME} held against {FACTION}"}
	msg := EventMessage{
		Kind:        EventKindDefend,
		Transition:  EventTransitionSucceeded,
		DefendEvent: fixedDefendEvent(3, EnemyCyborg),
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Region 3 for Cyborg = "Pictor Sector"
	if !strings.Contains(result, "Pictor Sector") {
		t.Errorf("expected region name, got: %s", result)
	}
	if !strings.Contains(result, "Cyborg") {
		t.Errorf("expected faction name, got: %s", result)
	}
}

func TestRenderEvent_DefendFailed(t *testing.T) {
	templates := Templates{DefendRegionFailed: "{REGION_NAME} fell to {FACTION}"}
	msg := EventMessage{
		Kind:        EventKindDefend,
		Transition:  EventTransitionFailed,
		DefendEvent: fixedDefendEvent(2, EnemyBug),
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Region 2 for Bug = "Kruger System"
	if !strings.Contains(result, "Kruger System") {
		t.Errorf("expected region name, got: %s", result)
	}
}

func TestRenderEvent_AttackHomeworldStarted(t *testing.T) {
	templates := Templates{AttackHomeworldStarted: "attacking {FACTION} homeworld ends {END_TIME_FORMATTED}"}
	msg := EventMessage{
		Kind:        EventKindAttack,
		Transition:  EventTransitionStarted,
		AttackEvent: fixedAttackEvent(EnemyCyborg),
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Cyborg") {
		t.Errorf("expected faction name, got: %s", result)
	}
}

func TestRenderEvent_AttackSucceeded(t *testing.T) {
	templates := Templates{AttackSucceeded: "{FACTION} defeated"}
	msg := EventMessage{
		Kind:        EventKindAttack,
		Transition:  EventTransitionSucceeded,
		AttackEvent: fixedAttackEvent(EnemyIlluminate),
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Illuminate") {
		t.Errorf("expected faction name, got: %s", result)
	}
}

func TestRenderEvent_AttackFailed(t *testing.T) {
	templates := Templates{AttackFailed: "{FACTION} defended homeworld"}
	msg := EventMessage{
		Kind:        EventKindAttack,
		Transition:  EventTransitionFailed,
		AttackEvent: fixedAttackEvent(EnemyBug),
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Bug") {
		t.Errorf("expected faction name, got: %s", result)
	}
}

func TestRenderEvent_NilDefendEvent(t *testing.T) {
	templates := Templates{DefendRegionStarted: "{FACTION}"}
	msg := EventMessage{Kind: EventKindDefend, Transition: EventTransitionStarted}
	_, err := RenderEvent(templates, msg, identityFormatter)
	if err == nil {
		t.Error("expected error for nil defend event, got nil")
	}
}

func TestRenderEvent_NilAttackEvent(t *testing.T) {
	templates := Templates{AttackHomeworldStarted: "{FACTION}"}
	msg := EventMessage{Kind: EventKindAttack, Transition: EventTransitionStarted}
	_, err := RenderEvent(templates, msg, identityFormatter)
	if err == nil {
		t.Error("expected error for nil attack event, got nil")
	}
}

func TestRenderEvent_UnknownKind(t *testing.T) {
	msg := EventMessage{Kind: "unknown", Transition: EventTransitionStarted}
	_, err := RenderEvent(Templates{}, msg, identityFormatter)
	if err == nil {
		t.Error("expected error for unknown kind, got nil")
	}
}

func TestRenderEvent_PlayersVariable(t *testing.T) {
	templates := Templates{DefendRegionStarted: "players: {PLAYERS}"}
	msg := EventMessage{
		Kind:        EventKindDefend,
		Transition:  EventTransitionStarted,
		DefendEvent: fixedDefendEvent(1, EnemyCyborg),
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "100") {
		t.Errorf("expected players count, got: %s", result)
	}
}

func TestRenderEvent_UnixTimestampVariable(t *testing.T) {
	templates := Templates{DefendRegionStarted: "{END_TIME_UNIX}"}
	e := fixedDefendEvent(1, EnemyCyborg)
	msg := EventMessage{
		Kind:        EventKindDefend,
		Transition:  EventTransitionStarted,
		DefendEvent: e,
	}
	result, err := RenderEvent(templates, msg, identityFormatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "1784674741" // fixedTime + 48h
	if result != expected {
		t.Errorf("expected unix timestamp %s, got: %s", expected, result)
	}
}
