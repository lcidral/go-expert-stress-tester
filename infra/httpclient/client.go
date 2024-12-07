package httpclient

import (
	"go-expert-stress-test/domain"
	"net/http"
	"time"
)

type Client struct {
	client *http.Client
}

// instancia um novo cliente com timeout de 30s
func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// executa uma requisicao GET e retorna TestResult
func (c *Client) Get(url string) (*domain.TestResult, error) {
	start := time.Now()

	resp, err := c.client.Get(url)
	duration := time.Since(start)

	if err != nil {
		return &domain.TestResult{
			Duration: duration,
			Error:    err,
		}, nil
	}
	defer resp.Body.Close()

	return &domain.TestResult{
		Duration: duration,
		Status:   resp.StatusCode,
	}, nil
}
