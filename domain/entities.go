package domain

import "time"

type TestConfig struct {
	URL         string // url que será testada
	Requests    int    // numero de requests que serao enviados
	Concurrency int    // numero de workers que serao usados para enviar as requisicoes
}

type TestResult struct {
	Duration time.Duration // duracao do teste
	Status   int
	Error    error
}

type TestReport struct {
	TotalDuration   time.Duration // duracao total do teste
	TotalRequests   int           // total de requisicoes disparadas contra o alvo
	SuccessRequests int           // requisicoes com sucesso
	StatusDistrib   map[int]int   // map de inteiros que armazens o resultado entre HTTP Status Code
	ErrorCount      int           // numero de erros
	AverageDuration time.Duration // duracao media de uma requisicao
}
