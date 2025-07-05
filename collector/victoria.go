package collector

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type VictoriaClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewVictoriaClient(baseURL string, logger *slog.Logger) *VictoriaClient {
	return &VictoriaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (v *VictoriaClient) SendMetrics(ctx context.Context, metrics map[string]float64) error {
	if len(metrics) == 0 {
		v.logger.DebugContext(ctx, "No metrics to send")
		return nil
	}

	// Convert metrics to Prometheus format
	var lines []string
	for name, value := range metrics {
		line := fmt.Sprintf("%s %f", name, value)
		lines = append(lines, line)
	}

	data := strings.Join(lines, "\n")

	v.logger.DebugContext(ctx, "Sending metrics to VictoriaMetrics",
		"count", len(metrics),
		"data", data,
		"url", v.baseURL)

	return v.sendData(ctx, data)
}

// Ping checks if VictoriaMetrics is accessible
func (v *VictoriaClient) Ping(ctx context.Context) error {
	url := v.baseURL + "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating ping request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("VictoriaMetrics health check failed: %s", resp.Status)
	}

	v.logger.InfoContext(ctx, "VictoriaMetrics is healthy")
	return nil
}

func (v *VictoriaClient) sendData(ctx context.Context, data string) error {
	url := v.baseURL + "/api/v1/import/prometheus"

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("VictoriaMetrics returned status %d: %s", resp.StatusCode, resp.Status)
	}

	v.logger.DebugContext(ctx, "Successfully sent metrics", "status", resp.Status)
	return nil
}
