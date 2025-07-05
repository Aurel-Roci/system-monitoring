package collector

import (
	"os"
	"strconv"
	"strings"

	"system-monitoring/models"
)

func GetLoadStats() (*models.LoadStats, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return &models.LoadStats{}, err
	}
	line := strings.Split(string(data), "\n")[0]

	fields := strings.Fields(line)

	oneMinAvg, _ := strconv.ParseFloat(fields[0], 64)
	fiveMinAvg, _ := strconv.ParseFloat(fields[1], 64)
	fifteenMinAvg, _ := strconv.ParseFloat(fields[2], 64)

	return &models.LoadStats{
		OneMinAvg:     oneMinAvg,
		FiveMinAvg:    fiveMinAvg,
		FifteenMinAvg: fifteenMinAvg,
	}, nil
}
