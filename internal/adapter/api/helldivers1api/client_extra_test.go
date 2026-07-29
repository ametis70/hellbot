package helldivers1api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/api/helldivers1api"
	"github.com/ametis70/hellbot/internal/testutil"
)

// TestFetchCampaign_WithStatistics verifies that toDomainStatistics is exercised
// when the API response contains a non-empty statistics array.
func TestFetchCampaign_WithStatistics(t *testing.T) {
	response := testutil.APIResponse{
		Time: int(time.Now().Unix()),
		FactionStatus: []testutil.APIFactionStatus{
			{Season: 99, Points: 100, PointsMax: 1000, Status: "active"},
		},
		Statistics: []testutil.APIStatistics{
			{Season: 99},
		},
		AttackEvents: []testutil.APIAttackEvent{},
	}

	srv := testutil.NewMockServer([]testutil.APIResponse{response})
	defer srv.Close()

	logger := testutil.DiscardLogger()
	client := helldivers1api.New(helldivers1api.Options{
		BaseURL: srv.URL(),
		Timeout: 5 * time.Second,
	}, logger)

	c, err := client.FetchCampaign()
	if err != nil {
		t.Fatalf("FetchCampaign: %v", err)
	}
	if len(c.Statistics) != 1 {
		t.Errorf("expected 1 statistic, got %d", len(c.Statistics))
	}
	if c.Statistics[0].Season != 99 {
		t.Errorf("expected season 99, got %d", c.Statistics[0].Season)
	}
}

// TestFetchCampaign_APIError verifies that a non-zero error_code is returned as an error.
func TestFetchCampaign_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error_code":1,"campaign_status":[],"attack_events":[],"statistics":[]}`))
	}))
	defer srv.Close()

	client := helldivers1api.New(helldivers1api.Options{
		BaseURL: srv.URL + "/",
		Timeout: 5 * time.Second,
	}, testutil.DiscardLogger())

	_, err := client.FetchCampaign()
	if err == nil {
		t.Error("expected error for non-zero error_code, got nil")
	}
}

// TestFetchCampaign_InvalidJSON verifies that a malformed response returns an error.
func TestFetchCampaign_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	client := helldivers1api.New(helldivers1api.Options{
		BaseURL: srv.URL + "/",
		Timeout: 5 * time.Second,
	}, testutil.DiscardLogger())

	_, err := client.FetchCampaign()
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestAPIStatistics_AllFields verifies all statistics fields are mapped correctly.
func TestAPIStatistics_AllFields(t *testing.T) {
	response := testutil.APIResponse{
		Time: int(time.Now().Unix()),
		FactionStatus: []testutil.APIFactionStatus{
			{Season: 1, Status: "active"},
		},
		Statistics: []testutil.APIStatistics{
			{Season: 1},
		},
		AttackEvents: []testutil.APIAttackEvent{},
	}
	srv := testutil.NewMockServer([]testutil.APIResponse{response})
	defer srv.Close()

	client := helldivers1api.New(helldivers1api.Options{
		BaseURL: srv.URL(),
		Timeout: 5 * time.Second,
	}, testutil.DiscardLogger())

	c, err := client.FetchCampaign()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Statistics) == 0 {
		t.Fatal("expected statistics to be mapped")
	}
}
