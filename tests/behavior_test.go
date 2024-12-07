package tests

import (
	"go-expert-stress-test/domain"
	"go-expert-stress-test/tests/mocks"
	"go-expert-stress-test/usecases"
	"testing"
	"time"
)

func TestRequestDistribution(t *testing.T) {
	tests := []struct {
		name        string
		config      domain.TestConfig
		results     []domain.TestResult
		expectError bool
	}{
		{
			name: "Deve respeitar limite de concorrência",
			config: domain.TestConfig{
				URL:         "http://test.com",
				Requests:    100,
				Concurrency: 10,
			},
			results: []domain.TestResult{
				{Duration: 10 * time.Millisecond, Status: 200},
			},
		},
		{
			name: "Deve completar todos os requests mesmo com erros",
			config: domain.TestConfig{
				URL:         "http://test.com",
				Requests:    50,
				Concurrency: 5,
			},
			results: []domain.TestResult{
				{Duration: 10 * time.Millisecond, Status: 200},
				{Duration: 10 * time.Millisecond, Status: 500},
				{Duration: 10 * time.Millisecond, Error: domain.ErrTimeout},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := mocks.NewMockHTTPClientWithMetrics(tt.results)
			loadTester := usecases.NewLoadTesterUseCase(mockClient, usecases.NewReporter())

			// Execute
			report, err := loadTester.Execute(tt.config)

			// Validate error expectation
			if (err != nil) != tt.expectError {
				t.Errorf("Erro inesperado: %v", err)
			}

			// Get metrics
			totalCalls, maxConcurrency, uniqueURLs := mockClient.GetMetrics()

			// Validate total requests
			if totalCalls != tt.config.Requests {
				t.Errorf("Número total de requests incorreto: got %d, want %d",
					totalCalls, tt.config.Requests)
			}

			// Validate concurrency
			if maxConcurrency > tt.config.Concurrency {
				t.Errorf("Concorrência máxima excedida: got %d, want <= %d",
					maxConcurrency, tt.config.Concurrency)
			}

			// Validate URL consistency
			if uniqueURLs != 1 {
				t.Errorf("Múltiplas URLs detectadas: got %d, want 1", uniqueURLs)
			}

			// Validate report totals match
			if report != nil && report.TotalRequests != tt.config.Requests {
				t.Errorf("Total de requests no relatório incorreto: got %d, want %d",
					report.TotalRequests, tt.config.Requests)
			}
		})
	}
}

func TestConcurrencyControl(t *testing.T) {
	config := domain.TestConfig{
		URL:         "http://test.com",
		Requests:    1000,
		Concurrency: 50,
	}

	mockClient := mocks.NewMockHTTPClientWithMetrics([]domain.TestResult{
		{Duration: 10 * time.Millisecond, Status: 200},
	})

	loadTester := usecases.NewLoadTesterUseCase(mockClient, usecases.NewReporter())

	// Execute test with high concurrency
	start := time.Now()
	_, err := loadTester.Execute(config)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	totalCalls, maxConcurrency, _ := mockClient.GetMetrics()

	// Validate metrics
	if totalCalls != config.Requests {
		t.Errorf("Total de requests incorreto: got %d, want %d",
			totalCalls, config.Requests)
	}

	if maxConcurrency > config.Concurrency {
		t.Errorf("Limite de concorrência excedido: got %d, want <= %d",
			maxConcurrency, config.Concurrency)
	}

	// Validate reasonable execution time
	expectedMinDuration := time.Duration(config.Requests/config.Concurrency) * 10 * time.Millisecond
	if duration < expectedMinDuration {
		t.Errorf("Execução muito rápida, possível problema no controle de concorrência")
	}
}

func TestRequestCompleteness(t *testing.T) {
	results := []domain.TestResult{
		{Duration: 10 * time.Millisecond, Status: 200},
		{Duration: 10 * time.Millisecond, Error: domain.ErrTimeout},
		{Duration: 10 * time.Millisecond, Status: 500},
		{Duration: 10 * time.Millisecond, Error: domain.ErrConnection},
	}

	config := domain.TestConfig{
		URL:         "http://test.com",
		Requests:    100,
		Concurrency: 10,
	}

	mockClient := mocks.NewMockHTTPClientWithMetrics(results)
	loadTester := usecases.NewLoadTesterUseCase(mockClient, usecases.NewReporter())

	report, err := loadTester.Execute(config)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	totalCalls, _, _ := mockClient.GetMetrics()

	// Validate total requests completed
	if totalCalls != config.Requests {
		t.Errorf("Requests incompletos: got %d, want %d",
			totalCalls, config.Requests)
	}

	// Validate error reporting
	totalResults := 0
	for _, count := range report.StatusDistrib {
		totalResults += count
	}
	totalResults += report.ErrorCount

	if totalResults != config.Requests {
		t.Errorf("Resultados perdidos no relatório: got %d, want %d",
			totalResults, config.Requests)
	}
}
