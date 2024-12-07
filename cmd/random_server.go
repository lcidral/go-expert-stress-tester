package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type StatusWeight struct {
	Code   int
	Weight int
}

var (
	statusCodes = []int{
		http.StatusOK,                  // 200
		http.StatusCreated,             // 201
		http.StatusAccepted,            // 202
		http.StatusBadRequest,          // 400
		http.StatusUnauthorized,        // 401
		http.StatusForbidden,           // 403
		http.StatusNotFound,            // 404
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusServiceUnavailable,  // 503
	}

	weights = []int{
		60, // 200 - 60% chance
		5,  // 201
		5,  // 202
		5,  // 400
		5,  // 401
		3,  // 403
		5,  // 404
		2,  // 429
		5,  // 500
		5,  // 503
	}
)

type RequestCounter struct {
	mu            sync.Mutex
	count         int
	lastRequest   time.Time
	shutdownTimer *time.Timer
}

type Response struct {
	Status     int           `json:"status"`
	Delay      time.Duration `json:"delay"`
	RequestNum int           `json:"request_number"`
}

func NewRequestCounter() *RequestCounter {
	return &RequestCounter{
		lastRequest: time.Now(),
	}
}

func (rc *RequestCounter) Increment() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.count++
	rc.lastRequest = time.Now()

	// após 100 requisições, monitora o container e desliga
	if rc.count > 100 {
		if rc.shutdownTimer == nil {
			rc.shutdownTimer = time.AfterFunc(5*time.Second, func() {
				log.Printf("Inatividade detectada após %d requisições. Desligando servidor...", rc.count)
				os.Exit(0)
			})
		} else {
			rc.shutdownTimer.Reset(5 * time.Second)
		}
	}

	return rc.count
}

func selectWeightedStatusCode(weights []StatusWeight) int {
	total := 0
	for _, w := range weights {
		total += w.Weight
	}

	r := rand.Intn(total)
	for _, w := range weights {
		r -= w.Weight
		if r < 0 {
			return w.Code
		}
	}

	return http.StatusOK
}

func main() {
	// configuração da porta, se nao houve valor externos, usa padrao 8080
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// contador de requisições
	counter := NewRequestCounter()

	// setup do servidor
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      http.DefaultServeMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// carregas configuracoes - pesos por http code
	statusWeights := parseStatusWeights(os.Getenv("STATUS_WEIGHTS"))

	// rota principal
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Gera delay aleatório entre 100ms e 2s
		delay := time.Duration(100+rand.Intn(1900)) * time.Millisecond
		time.Sleep(delay)

		// obtem um status code com base nas configuraçoes de peso que foram passadas no compose.yaml
		statusCode := selectWeightedStatusCode(statusWeights)
		requestNum := counter.Increment()

		// cabeçalhos da resposta http
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Response-Delay", delay.String())
		w.Header().Set("X-Response-Time", time.Now().Format(time.RFC3339))
		w.Header().Set("X-Request-Number", strconv.Itoa(requestNum))

		// escreve o status code na resposta
		w.WriteHeader(statusCode)

		// prepara response
		response := Response{
			Status:     statusCode,
			Delay:      delay,
			RequestNum: requestNum,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Erro ao codificar resposta: %v", err)
		}

		// escreve um log das requisicoes obtidas
		log.Printf("Request #%d: Status %d, Delay %v", requestNum, statusCode, delay)
	})

	// rota padrao para verificar se o container está em execucao
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy"}`)
	})

	// cria um canal para monitorar o sinal de término do container
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// inicia o servisor em uma goroutine
	go func() {
		log.Printf("Servidor iniciado em http://localhost%s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	// aguarda pelo final de termino
	<-done
	log.Println("Servidor recebeu sinal de término...")

	// cria um contexto de 10 segundos para forçar o desligamento
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// gracefully shutdown :)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Erro ao desligar servidor: %v", err)
	}

	log.Println("Servidor finalizado com sucesso")
}

func parseStatusWeights(weightsStr string) []StatusWeight {
	if weightsStr == "" {
		// Retorna configuração padrão se nenhuma for fornecida
		return []StatusWeight{
			{Code: http.StatusOK, Weight: 60},
			{Code: http.StatusCreated, Weight: 5},
			{Code: http.StatusAccepted, Weight: 5},
			{Code: http.StatusBadRequest, Weight: 5},
			{Code: http.StatusUnauthorized, Weight: 5},
			{Code: http.StatusForbidden, Weight: 3},
			{Code: http.StatusNotFound, Weight: 5},
			{Code: http.StatusTooManyRequests, Weight: 2},
			{Code: http.StatusInternalServerError, Weight: 5},
			{Code: http.StatusServiceUnavailable, Weight: 5},
		}
	}

	var weights []StatusWeight
	pairs := strings.Split(weightsStr, ",")

	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}

		code, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		weight, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		weights = append(weights, StatusWeight{Code: code, Weight: weight})
	}

	return weights
}
