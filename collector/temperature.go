package collector

import (
	"os"
	"strconv"
	"strings"
	"time"

	"system-monitoring/models"
)

func GetTempStats() (*models.SystemTemp, error) {

	dat, e := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if e != nil {
		return &models.SystemTemp{}, e
	}
	tempStr := strings.TrimSpace(string(dat))

	tempMilliC, err := strconv.Atoi(tempStr)
	if err != nil {
		return &models.SystemTemp{}, err
	}
	tempC := float64(tempMilliC) / 1000.0

	temp := models.SystemTemp{
		TempInC:   tempC,
		Timestamp: time.Now(),
	}

	return &temp, nil
}
