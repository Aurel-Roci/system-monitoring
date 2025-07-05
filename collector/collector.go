package collector

import (
	"context"
	"log/slog"

	"system-monitoring/models"
)

type Collector struct {
	name    string
	enabled bool
	fn      func() (interface{}, error)
}

func CollectAndSendMetrics(ctx context.Context, config *models.Config, logger *slog.Logger) error {
	slog.Info("Starting concurrent collection...")

	vmClient := NewVictoriaClient(config.GetVictoriaMetricsURL(), logger)

	if err := vmClient.Ping(ctx); err != nil {
		logger.Error("VictoriaMetrics is not accessible", "error", err)
		return err
	}

	allMetrics, err := CollectAllMetrics(config, logger)
	if err != nil {
		return err
	}

	if len(allMetrics) > 0 {
		err = vmClient.SendMetrics(ctx, allMetrics)
		if err != nil {
			logger.Error("Failed to send metrics to VictoriaMetrics", "error", err)
			return err
		}
		logger.Info("Successfully sent metrics to VictoriaMetrics", "count", len(allMetrics))
	} else {
		logger.Warn("No metrics collected, nothing to send")
	}

	slog.Info("All results collected!")
	return nil
}

func CollectAllMetrics(config *models.Config, logger *slog.Logger) (map[string]float64, error) {
	logger.Debug("Starting metric collection...")

	resultChan := make(chan models.ResultPtr)
	allMetrics := make(map[string]float64)

	collectors := []Collector{
		{"temperature", config.Monitoring.EnableTemperature, func() (interface{}, error) { return GetTempStats() }},
		{"memory", config.Monitoring.EnableMemory, func() (interface{}, error) { return GetMemoryStats() }},
		{"cpu", config.Monitoring.EnableCPU, func() (interface{}, error) { return GetCpuStats() }},
		{"load", config.Monitoring.EnableLoad, func() (interface{}, error) { return GetLoadStats() }},
	}

	enabledCount := 0
	for _, collector := range collectors {
		if collector.enabled {
			enabledCount++
			go func(c Collector) {
				data, err := c.fn()
				resultChan <- models.ResultPtr{Type: c.name, Data: data, Error: err}
			}(collector)
		} else {
			logger.Debug("Collector disabled, skipping", "collector", collector.name)
		}
	}

	if enabledCount == 0 {
		logger.Warn("No collectors enabled")
		return allMetrics, nil
	}

	// Collect results from enabled collectors
	for i := 0; i < enabledCount; i++ {
		result := <-resultChan
		if result.Error != nil {
			logger.Error("Error collecting metrics", "type", result.Type, "error", result.Error)
			continue
		}

		metrics := convertToMetrics(result)
		for name, value := range metrics {
			allMetrics[name] = value
		}

		logger.Debug("Metrics collected", "type", result.Type, "count", len(metrics))
	}

	logger.Info("All metrics collected", "total_count", len(allMetrics))
	return allMetrics, nil
}

func convertToMetrics(result models.ResultPtr) map[string]float64 {
	metrics := make(map[string]float64)

	switch result.Type {
	case "temperature":
		if temp, ok := result.Data.(*models.SystemTemp); ok {
			metrics["system_temperature_celsius"] = temp.TempInC
		}
	case "memory":
		if memory, ok := result.Data.(*models.MemoryStats); ok {
			// Convert KB to MB and calculate percentages
			totalMB := float64(memory.Total) / 1024
			availableMB := float64(memory.Available) / 1024
			freeMB := float64(memory.Free) / 1024
			usedMB := totalMB - availableMB

			metrics["memory_total_mb"] = totalMB
			metrics["memory_available_mb"] = availableMB
			metrics["memory_free_mb"] = freeMB
			metrics["memory_used_mb"] = usedMB
			metrics["memory_usage_percent"] = (usedMB / totalMB) * 100
		}
	case "cpu":
		if cpu, ok := result.Data.(*models.CPUStats); ok {
			metrics["cpu_usage_percent"] = cpu.UsagePct
		}
	case "load":
		if load, ok := result.Data.(*models.LoadStats); ok {
			metrics["load_average_1min"] = load.OneMinAvg
			metrics["load_average_5min"] = load.FiveMinAvg
			metrics["load_average_15min"] = load.FifteenMinAvg
		}
	}

	return metrics
}
