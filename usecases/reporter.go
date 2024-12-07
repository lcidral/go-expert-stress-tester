package usecases

import (
	"go-expert-stress-test/domain"
	"time"
)

type Reporter struct{}

func NewReporter() *Reporter {
	return &Reporter{}
}

// gera um report para o resultado do teste com a duração e o status
func (r *Reporter) GenerateReport(results []domain.TestResult, totalDuration time.Duration) *domain.TestReport {
	report := &domain.TestReport{
		TotalDuration: totalDuration,
		TotalRequests: len(results),
		StatusDistrib: make(map[int]int),
	}

	var totalReqDuration time.Duration
	for _, result := range results {
		totalReqDuration += result.Duration

		if result.Error != nil {
			report.ErrorCount++
			continue
		}

		report.StatusDistrib[result.Status]++
		if result.Status == 200 {
			report.SuccessRequests++
		}
	}

	if report.TotalRequests > 0 {
		report.AverageDuration = totalReqDuration / time.Duration(report.TotalRequests)
	}

	return report
}
