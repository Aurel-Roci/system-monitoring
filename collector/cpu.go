package collector

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"system-monitoring/models"
)

func readCPUUsage() (int64, int64, error) {

	dat, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(string(dat), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") { //Space is there since this gets all CPU cores
			fields := strings.Fields(line)

			user, _ := strconv.ParseInt(fields[1], 10, 64)
			nice, _ := strconv.ParseInt(fields[2], 10, 64)
			system, _ := strconv.ParseInt(fields[3], 10, 64)
			idle, _ := strconv.ParseInt(fields[4], 10, 64)
			iowait, _ := strconv.ParseInt(fields[5], 10, 64)

			totalTime := user + nice + system + idle + iowait
			busyTime := user + nice + system + iowait

			return totalTime, busyTime, nil
		}
	}
	return 0, 0, errors.New("could not find CPU line in /proc/stat")
}

func GetCpuStats() (*models.CPUStats, error) {

	total1, busy1, _ := readCPUUsage()

	time.Sleep(1 * time.Second)

	total2, busy2, _ := readCPUUsage()

	busyDiff := busy2 - busy1
	totalDiff := total2 - total1

	usage := float64(busyDiff) / float64(totalDiff) * 100

	return &models.CPUStats{
		UsagePct:  usage,
		CoreID:    -1,
		Timestamp: time.Now(),
	}, nil
}
