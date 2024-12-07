package mocks

import (
	"go-expert-stress-test/domain"
	"sync"
	"time"
)

// armazena metricas detalhadas das chamadas
type MockHTTPClientWithMetrics struct {
	mu                sync.Mutex
	calls             []time.Time
	activeConnections int
	maxConnections    int
	uniqueURLs        map[string]struct{}
	results           []domain.TestResult
}

// cria um mock do cliente http com os resultados
func NewMockHTTPClientWithMetrics(results []domain.TestResult) *MockHTTPClientWithMetrics {
	return &MockHTTPClientWithMetrics{
		calls:      make([]time.Time, 0),
		uniqueURLs: make(map[string]struct{}),
		results:    results,
	}
}

func (m *MockHTTPClientWithMetrics) Get(url string) (*domain.TestResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, time.Now())
	m.uniqueURLs[url] = struct{}{}

	m.activeConnections++
	if m.activeConnections > m.maxConnections {
		m.maxConnections = m.activeConnections
	}

	// simula algum trabalho no backend
	time.Sleep(10 * time.Millisecond)
	m.activeConnections--

	result := m.results[len(m.calls)%len(m.results)]
	return &result, nil
}

// retorna 3 valores conhecidos, numero de chamadas, numero de conxoes e url unicas
func (m *MockHTTPClientWithMetrics) GetMetrics() (totalCalls int, maxConcurrency int, uniqueURLs int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.calls), m.maxConnections, len(m.uniqueURLs)
}
