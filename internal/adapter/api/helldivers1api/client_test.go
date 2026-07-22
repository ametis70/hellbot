package helldivers1api

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stderr, nil))

const validCampaignJSON = `{
	"time": 1784505880,
	"error_code": 0,
	"campaign_status": [
		{
			"season": 159,
			"points": 280970,
			"points_taken": 604351,
			"points_max": 280970,
			"status": "defeated",
			"introduction_order": 2
		}
	],
	"defend_event": {
		"season": 159,
		"event_id": 5080,
		"start_time": 1784501941,
		"end_time": 1784674741,
		"region": 0,
		"enemy": 1,
		"points_max": 31602,
		"points": 486,
		"status": "active",
		"players_at_start": 0
	},
	"attack_events": [
		{
			"season": 159,
			"event_id": 924,
			"start_time": 1784291521,
			"end_time": 1784456642,
			"enemy": 0,
			"points_max": 31576,
			"points": 31576,
			"status": "success",
			"players_at_start": 184,
			"max_event_id": 924
		}
	],
	"statistics": []
}`

const apiErrorJSON = `{
	"time": 1784505880,
	"error_code": 2,
	"error_message": "Invalid action"
}`

func newTestClient(serverURL string) *Client {
	opts := DefaultOptions()
	opts.BaseURL = serverURL
	return New(opts, testLogger)
}

func TestFetchCampaign_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(validCampaignJSON))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	campaign, err := client.FetchCampaign()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if campaign.DefendEvent == nil {
		t.Fatal("expected defend event, got nil")
	}
	if campaign.DefendEvent.ID != 5080 {
		t.Errorf("expected defend event ID 5080, got %d", campaign.DefendEvent.ID)
	}
	if campaign.DefendEvent.Enemy != domain.EnemyIlluminate {
		t.Errorf("expected enemy Illuminate, got %v", campaign.DefendEvent.Enemy)
	}
	if len(campaign.AttackEvents) != 1 {
		t.Errorf("expected 1 attack event, got %d", len(campaign.AttackEvents))
	}
	if len(campaign.FactionsStatus) != 1 {
		t.Errorf("expected 1 faction, got %d", len(campaign.FactionsStatus))
	}
}

func TestFetchCampaign_TimestampsConverted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(validCampaignJSON))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	campaign, err := client.FetchCampaign()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedTime := time.Unix(1784505880, 0).UTC()
	if !campaign.Time.Equal(expectedTime) {
		t.Errorf("expected campaign time %v, got %v", expectedTime, campaign.Time)
	}

	expectedStart := time.Unix(1784501941, 0).UTC()
	if !campaign.DefendEvent.StartTime.Equal(expectedStart) {
		t.Errorf("expected defend start time %v, got %v", expectedStart, campaign.DefendEvent.StartTime)
	}
}

func TestFetchCampaign_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(apiErrorJSON))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.FetchCampaign()
	if err == nil {
		t.Error("expected error for API error code, got nil")
	}
}

func TestFetchCampaign_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.FetchCampaign()
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestFetchCampaign_ServerUnreachable(t *testing.T) {
	opts := DefaultOptions()
	opts.BaseURL = "http://127.0.0.1:1"
	opts.Timeout = 1 * time.Second
	client := New(opts, testLogger)

	_, err := client.FetchCampaign()
	if err == nil {
		t.Error("expected error for unreachable server, got nil")
	}
}

func TestFetchCampaign_NoDefendEvent(t *testing.T) {
	noDefendJSON := `{
		"time": 1784505880,
		"error_code": 0,
		"campaign_status": [],
		"defend_event": {"event_id": 0},
		"attack_events": [],
		"statistics": []
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(noDefendJSON))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	campaign, err := client.FetchCampaign()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if campaign.DefendEvent != nil {
		t.Errorf("expected nil defend event when ID is 0, got %+v", campaign.DefendEvent)
	}
}
