package main

import (
	"flag"
	"go-expert-stress-test/domain"
	"go-expert-stress-test/infra/httpclient"
	"go-expert-stress-test/interfaces/cli"
	"go-expert-stress-test/usecases"
	"log"
)

func main() {
	config := domain.TestConfig{}

	// attribui os argumentos ao config
	flag.StringVar(&config.URL, "url", "", "URL do serviço a ser testado")
	flag.IntVar(&config.Requests, "requests", 0, "Número total de requests")
	flag.IntVar(&config.Concurrency, "concurrency", 1, "Número de chamadas simultâneas")
	flag.Parse()

	// minima validação
	if config.URL == "" || config.Requests <= 0 {
		log.Fatal("URL e número de requests são obrigatórios")
	}

	// inicializa o client http
	httpClient := httpclient.NewClient()
	// inicializa o reporter que irá imprimir o resultado do teste
	reporter := usecases.NewReporter()
	// inicializa o loadTester com o HTTPClient e o Reporter
	loadTester := usecases.NewLoadTesterUseCase(httpClient, reporter)

	// executa o teste de carga
	report, err := loadTester.Execute(config)
	if err != nil {
		log.Fatal(err)
	}

	// inicializa o reporter
	presenter := cli.NewReportPresenter()

	// imprime o resultado do teste de carga
	presenter.Present(report)
}
