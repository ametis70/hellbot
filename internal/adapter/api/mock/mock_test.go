package mock_test

import (
	"testing"

	"github.com/ametis70/hellbot/internal/adapter/api/mock"
	"github.com/ametis70/hellbot/internal/testutil"
)

func TestMockFetcher_ServesFrames(t *testing.T) {
	f := mock.New(func() {}, testutil.DiscardLogger())

	// First fetch should succeed.
	c, err := f.FetchCampaign()
	if err != nil {
		t.Fatalf("expected no error on first fetch, got %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil campaign on first fetch")
	}
}

func TestMockFetcher_CancelsOnExhaustion(t *testing.T) {
	cancelled := false
	cancel := func() { cancelled = true }
	f := mock.New(cancel, testutil.DiscardLogger())

	// Drain all frames.
	for {
		_, err := f.FetchCampaign()
		if err != nil {
			t.Fatalf("unexpected error mid-scenario: %v", err)
		}
		if cancelled {
			break
		}
	}

	if !cancelled {
		t.Error("expected cancel to be called after all frames served")
	}

	// Further calls should still return a non-nil campaign (last frame).
	c, err := f.FetchCampaign()
	if err != nil {
		t.Errorf("expected no error after exhaustion, got %v", err)
	}
	if c == nil {
		t.Error("expected last frame returned after exhaustion")
	}
}
