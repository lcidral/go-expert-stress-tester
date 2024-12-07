package usecases

import (
	"fmt"
	"go-expert-stress-test/domain"
	"go-expert-stress-test/infra/progress"
	"sync"
	"time"
)

type LoadTesterUseCase struct {
	httpClient domain.HTTPClient
	reporter   domain.Reporter
}

// retorna instancia do LoadTesterUseCase
func NewLoadTesterUseCase(client domain.HTTPClient, reporter domain.Reporter) *LoadTesterUseCase {
	return &LoadTesterUseCase{
		httpClient: client,
		reporter:   reporter,
	}
}

func (lt *LoadTesterUseCase) Execute(config domain.TestConfig) (*domain.TestReport, error) {
	results := make([]domain.TestResult, 0, config.Requests)
	resultsChan := make(chan domain.TestResult, config.Requests)
	var wg sync.WaitGroup

	startTime := time.Now()

	// cria uma instancia do ProgressTracker com as configuesções
	progressTracker := progress.NewProgressTracker(config.Concurrency, config.Requests)
	progressTracker.Start()
	defer progressTracker.Stop()

	// canal para controlar a conclusao das chamadas
	done := make(chan struct{})

	// distribui as requisicoes entre os workers
	requestsPerWorker := config.Requests / config.Concurrency
	remainder := config.Requests % config.Concurrency

	// eu amo isso de mais
	for workerID := 0; workerID < config.Concurrency; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			requests := requestsPerWorker
			if id == config.Concurrency-1 {
				requests += remainder
			}

			for i := 0; i < requests; i++ {
				result, _ := lt.httpClient.Get(config.URL)
				resultsChan <- *result
				progressTracker.IncrementWorker(id)
			}
		}(workerID)
	}

	// espera a conclusao da goroutine e fecha os canais
	go func() {
		wg.Wait()
		close(resultsChan)
		close(done)
	}()

	// itera entre os resultados que vieram do canal
	for result := range resultsChan {
		results = append(results, result)
	}

	// espera a conclusao de tudo
	<-done

	// faz uma pausa para garantir que todas as barra de progresso concluiram o trabalho
	time.Sleep(500 * time.Millisecond)

	// limpa a tela para imprimir o resultado
	fmt.Print("\033[H\033[2J")

	// chama o reporter para gerar o resultado do relatorio
	return lt.reporter.GenerateReport(results, time.Since(startTime)), nil
}
