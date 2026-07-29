package testutil

import (
	"io"
	"log/slog"
	"sync"

	"github.com/ametis70/hellbot/internal/domain"
)

// DiscardLogger returns a logger that discards all output. Useful in tests that
// don't care about log output.
func DiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// MockFetcher implements port.Fetcher

type MockFetcher struct {
	Campaign *domain.CampaignStatus
	Err      error
}

func (m *MockFetcher) FetchCampaign() (*domain.CampaignStatus, error) {
	return m.Campaign, m.Err
}

// MockNotifier implements port.Notifier — records all received messages

type MockNotifier struct {
	mu       sync.Mutex
	Messages []domain.EventMessage
}

func (m *MockNotifier) Notify(msg domain.EventMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, msg)
	return nil
}

func (m *MockNotifier) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = nil
}

func (m *MockNotifier) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Messages)
}

func (m *MockNotifier) First() *domain.EventMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Messages) == 0 {
		return nil
	}
	return &m.Messages[0]
}

func (m *MockNotifier) Last() *domain.EventMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Messages) == 0 {
		return nil
	}
	return &m.Messages[len(m.Messages)-1]
}
