package domain

import "time"

type LoadTester interface {
	Execute(config TestConfig) (*TestReport, error)
}

type HTTPClient interface {
	Get(url string) (*TestResult, error)
}

type Reporter interface {
	GenerateReport(results []TestResult, duration time.Duration) *TestReport
}

type ProgressTracker interface {
	Start()
	Stop()
	IncrementWorker(workerID int)
}
