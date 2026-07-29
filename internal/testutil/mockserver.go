package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// APIResponse is the wire format returned by the Helldivers 1 API for the
// "get_campaign_status" action. It mirrors the private apiCampaignStatus type
// in the helldivers1api package so that tests can construct responses without
// importing internal types.
type APIResponse struct {
	Time          int                `json:"time"`
	ErrorCode     int                `json:"error_code"`
	FactionStatus []APIFactionStatus `json:"campaign_status"`
	DefendEvent   APIDefendEvent     `json:"defend_event"`
	AttackEvents  []APIAttackEvent   `json:"attack_events"`
	Statistics    []APIStatistics    `json:"statistics"`
}

type APIFactionStatus struct {
	Season            int    `json:"season"`
	Points            int    `json:"points"`
	PointsTaken       int    `json:"points_taken"`
	PointsMax         int    `json:"points_max"`
	Status            string `json:"status"`
	IntroductionOrder int    `json:"introduction_order"`
}

type APIDefendEvent struct {
	Season         int    `json:"season"`
	ID             int    `json:"event_id"`
	StartTime      int    `json:"start_time"`
	EndTime        int    `json:"end_time"`
	Region         int    `json:"region"`
	Enemy          int    `json:"enemy"`
	PointsMax      int    `json:"points_max"`
	Points         int    `json:"points"`
	Status         string `json:"status"`
	PlayersAtStart int    `json:"players_at_start"`
}

type APIAttackEvent struct {
	Season         int    `json:"season"`
	ID             int    `json:"event_id"`
	StartTime      int    `json:"start_time"`
	EndTime        int    `json:"end_time"`
	Enemy          int    `json:"enemy"`
	PointsMax      int    `json:"points_max"`
	Points         int    `json:"points"`
	Status         string `json:"status"`
	PlayersAtStart int    `json:"players_at_start"`
	MaxEventID     int    `json:"max_event_id"`
}

type APIStatistics struct {
	Season                 int `json:"season"`
	SeasonDuration         int `json:"season_duration"`
	Enemy                  int `json:"enemy"`
	Players                int `json:"players"`
	TotalUniquePlayers     int `json:"total_unique_players"`
	Missions               int `json:"missions"`
	SuccessfulMissions     int `json:"successful_missions"`
	TotalMissionDifficulty int `json:"total_mission_difficulty"`
	CompletedPlanets       int `json:"completed_planets"`
	DefendEvents           int `json:"defend_events"`
	SuccessfulDefendEvents int `json:"successful_defend_events"`
	AttackEvents           int `json:"attack_events"`
	SuccessfulAttackEvents int `json:"successful_attack_events"`
	Deaths                 int `json:"deaths"`
	Kills                  int `json:"kills"`
	Accidentals            int `json:"accidentals"`
	Shots                  int `json:"shots"`
	Hits                   int `json:"hits"`
}

// MockServer is an httptest.Server that serves Helldivers 1 API responses from
// a pre-loaded sequence, advancing one step per request. When the sequence is
// exhausted it keeps returning the last response.
type MockServer struct {
	Server    *httptest.Server
	mu        sync.Mutex
	responses []APIResponse
	idx       int
}

// NewMockServer creates and starts a MockServer loaded with the given sequence
// of responses.
func NewMockServer(responses []APIResponse) *MockServer {
	ms := &MockServer{responses: responses}
	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handle))
	return ms
}

func (ms *MockServer) handle(w http.ResponseWriter, r *http.Request) {
	ms.mu.Lock()
	resp := ms.current()
	ms.advance()
	ms.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (ms *MockServer) current() APIResponse {
	if len(ms.responses) == 0 {
		return APIResponse{}
	}
	return ms.responses[ms.idx]
}

func (ms *MockServer) advance() {
	if ms.idx < len(ms.responses)-1 {
		ms.idx++
	}
}

// URL returns the base URL of the mock server (suitable for helldivers1api.Options.BaseURL).
func (ms *MockServer) URL() string {
	return ms.Server.URL + "/"
}

// Close shuts down the mock server.
func (ms *MockServer) Close() {
	ms.Server.Close()
}

// CallCount returns how many requests have been served.
func (ms *MockServer) CallCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.idx
}
