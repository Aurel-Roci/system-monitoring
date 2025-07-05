package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"system-monitoring/collector"
	"system-monitoring/models"
	"time"
)

func logEnabledMonitoring(config *models.Config) {
	enabledFeatures := []string{}

	if config.Monitoring.EnableCPU {
		enabledFeatures = append(enabledFeatures, "CPU")
	}
	if config.Monitoring.EnableMemory {
		enabledFeatures = append(enabledFeatures, "Memory")
	}
	if config.Monitoring.EnableTemperature {
		enabledFeatures = append(enabledFeatures, "Temperature")
	}
	if config.Monitoring.EnableLoad {
		enabledFeatures = append(enabledFeatures, "Load")
	}

	if len(enabledFeatures) == 0 {
		slog.Warn("No monitoring features enabled!")
	} else {
		slog.Info("Monitoring features enabled", "features", enabledFeatures)
	}
}

func main() {
	slog.Info("App starting...")

	config, err := models.LoadConfig()
	if err != nil {
		slog.Error("Error loading config", "error", err)
		return
	}
	slog.Debug("Config loaded", "config", config)

	if err := models.SetupLogger(config); err != nil {
		slog.Error("Failed to setup logger", "error", err)
		return
	}

	slog.Info("Starting system monitor",
		"collection_interval", config.CollectionInterval,
		"database_type", config.Database.Type,
		"victoria_url", config.GetVictoriaMetricsURL(),
	)

	logEnabledMonitoring(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the monitoring loop
	if err := runMonitoringLoop(ctx, config); err != nil {
		slog.Error("Monitoring loop failed", "error", err)
		os.Exit(1)
	}

	slog.Info("System monitor shutdown complete")
}

func runMonitoringLoop(ctx context.Context, config *models.Config) error {
	// Create a ticker for periodic collection
	ticker := time.NewTicker(config.CollectionInterval)
	defer ticker.Stop()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("Monitoring loop started", "interval", config.CollectionInterval)

	// Test VictoriaMetrics connection on startup
	if err := testVictoriaMetricsConnection(ctx, config); err != nil {
		slog.Error("VictoriaMetrics connection test failed", "error", err)
		slog.Info("Will continue trying to send metrics...")
	}

	// Main monitoring loop
	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping monitoring loop")
			return nil

		case sig := <-sigChan:
			slog.Info("Received shutdown signal", "signal", sig)
			return nil

		case <-ticker.C:
			if err := collectAndSendMetrics(ctx, config); err != nil {
				slog.Error("Metric collection failed", "error", err)
			}
		}
	}
}

func testVictoriaMetricsConnection(ctx context.Context, config *models.Config) error {
	vmClient := collector.NewVictoriaClient(config.GetVictoriaMetricsURL(), slog.Default())

	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return vmClient.Ping(testCtx)
}

func collectAndSendMetrics(ctx context.Context, config *models.Config) error {
	start := time.Now()

	// Use a timeout for the collection process
	collectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := collector.CollectAndSendMetrics(collectCtx, config, slog.Default())

	duration := time.Since(start)
	if err != nil {
		slog.Error("Failed to collect and send metrics",
			"error", err,
			"duration", duration)
		return err
	}

	slog.Debug("Metric collection cycle completed", "duration", duration)
	return nil
}
