package collector

import (
	"os"
	"strconv"
	"strings"
	"time"

	"system-monitoring/models"
)

func GetMemoryStats() (*models.MemoryStats, error) {
	dat, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return &models.MemoryStats{}, err
	}
	lines := strings.Split(string(dat), "\n")

	var memTotal, memFree, memAvailable int64

	for _, line := range lines {
		if strings.Contains(line, "MemTotal:") {
			numberStr := strings.Fields(line)[1]
			memTotal, err = strconv.ParseInt(numberStr, 10, 64)
			if err != nil {
				return &models.MemoryStats{}, err
			}
		}
		if strings.Contains(line, "MemFree:") {
			numberStr := strings.Fields(line)[1]
			memFree, err = strconv.ParseInt(numberStr, 10, 64)
			if err != nil {
				return &models.MemoryStats{}, err
			}
		}
		if strings.Contains(line, "MemAvailable:") {
			numberStr := strings.Fields(line)[1]
			memAvailable, err = strconv.ParseInt(numberStr, 10, 64)
			if err != nil {
				return &models.MemoryStats{}, err
			}
		}
	}
	return &models.MemoryStats{
		Total:     memTotal / 1000,
		Free:      memFree / 1000,
		Available: memAvailable / 1000,
		Timestamp: time.Now(),
	}, nil
}
