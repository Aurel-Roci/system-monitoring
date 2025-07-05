package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)


func main(){

	// tempStats, _ := getTempStats()
	// memStats, _ := getMemoryStats()
	// cpuStats, _ := getCpuStats()

	collectMetricsConcurrently()
}


type Result struct {
    Type  string
    Data  interface{}  // This can hold any type
    Error error
}

type LoadStats struct {
    OneMinAvg     float64
    FiveMinAvg    float64
    FifteenMinAvg float64
}

type MemoryStats struct { 
    Available int64
    Free int64
    Total int64
    Timestamp time.Time
}

type CPUStats struct {
    UsagePct  float64 
    CoreID int      
    Timestamp time.Time
}

type SystemTemp struct {
    TempInC   float64
    Timestamp time.Time
}


func collectMetricsConcurrently() {
    fmt.Println("Starting concurrent collection...")

	resultChan := make(chan Result) 

	collectors := []struct {
    name string
    fn   func() (interface{}, error)
	}{
		{"temperature", func() (interface{}, error) { return getTempStats() }},
		{"memory", func() (interface{}, error) { return getMemoryStats() }},
		{"cpu", func() (interface{}, error) { return getCpuStats() }},
		{"load", func() (interface{}, error) { return getLoadStats() }},
	}

    // Start goroutines
    for _, collector := range collectors {
		go func (name string, collectFn func()(interface{}, error)) {
			data, err := collectFn()
			resultChan <- Result{Type: name, Data: data, Error: err}
		}(collector.name, collector.fn)
	}
	
	for i := 0; i < len(collectors); i++ {
        result := <-resultChan
        
        if result.Error != nil {
            fmt.Printf("Error collecting %s: %v\n", result.Type, result.Error)
            continue
        }
        
        // Now we need to convert interface{} back to the right type
        switch result.Type {
			case "temperature":
				temp := result.Data.(SystemTemp)  // Type assertion
				fmt.Printf("Temperature: %.2f°C\n", temp.TempInC)
			case "memory":
				memory := result.Data.(MemoryStats)
				fmt.Printf("Memory Available: %d MB\n", memory.Available)
			case "cpu":
				cpu := result.Data.(CPUStats)
				fmt.Printf("CPU Usage: %.2f%%\n", cpu.UsagePct)
			case "load":
				load := result.Data.(LoadStats)
				fmt.Printf("Load Average: %.2f, %.2f, %.2f\n", load.OneMinAvg, load.FiveMinAvg, load.FifteenMinAvg)
        }
    }
    fmt.Println("All results collected!")
}

func getTempStats()(SystemTemp, error){

	dat, e := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if e != nil {
        return SystemTemp{}, e
    }
	tempStr := strings.TrimSpace(string(dat))

	tempMilliC, err := strconv.Atoi(tempStr)
	if err != nil {
		return SystemTemp{}, err
	}
	tempC := float64(tempMilliC) / 1000.0
	
	temp := SystemTemp{
		TempInC:   tempC,
		Timestamp: time.Now(),
	}

	return temp, nil
}

func getMemoryStats() (MemoryStats, error) {
    dat, err := os.ReadFile("/proc/meminfo")
    if err != nil {
        return MemoryStats{}, err
    }
	lines := strings.Split(string(dat), "\n")
    
	var memTotal, memFree, memAvailable int64

    for _, line := range lines {
        if strings.Contains(line, "MemTotal:") {
			numberStr := strings.Fields(line)[1]
			memTotal, err = strconv.ParseInt(numberStr, 10, 64)
            if err != nil {
                return MemoryStats{}, err
            }
        }
        if strings.Contains(line, "MemFree:") {
			numberStr := strings.Fields(line)[1]
			memFree, err = strconv.ParseInt(numberStr, 10, 64)
            if err != nil {
                return MemoryStats{}, err
            }
        }
        if strings.Contains(line, "MemAvailable:") {
			numberStr := strings.Fields(line)[1]
			memAvailable, err = strconv.ParseInt(numberStr, 10, 64)
            if err != nil {
                return MemoryStats{}, err
            }
        }
    }
	return MemoryStats{
        Total:     memTotal / 1000,
        Free:      memFree / 1000,
        Available: memAvailable / 1000,
        Timestamp: time.Now(),
    }, nil
}

func readCPUUsage() (int64, int64, error){
	
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
	return 0, 0, errors.New("couldn't read CPU usages") 
}

func getCpuStats() (CPUStats, error) {

	total1, busy1, _ := readCPUUsage()

    // fmt.Printf("First reading - Total: %d, Busy: %d\n", total1, busy1)

	time.Sleep(1 * time.Second)

	total2, busy2, _ := readCPUUsage()
    // fmt.Printf("Second reading - Total: %d, Busy: %d\n", total2, busy2)

	busyDiff := busy2 - busy1
    totalDiff := total2 - total1
    // fmt.Printf("Differences - Total diff: %d, Busy diff: %d\n", totalDiff, busyDiff)

	usage := float64(busyDiff) / float64(totalDiff) * 100

	return CPUStats{
		UsagePct: usage,
		CoreID: -1,
		Timestamp: time.Now(),
	}, nil
}

func getLoadStats() (LoadStats, error){
	data, err := os.ReadFile("/proc/loadavg")
	 if err != nil {
        return LoadStats{}, err
    }
	line := strings.Split(string(data), "\n")[0]

	fields := strings.Fields(line)

	oneMinAvg, _ := strconv.ParseFloat(fields[0], 64)
	fiveMinAvg, _ := strconv.ParseFloat(fields[1], 64)
	fifteenMinAvg, _ := strconv.ParseFloat(fields[2], 64)

	return LoadStats{
		OneMinAvg: oneMinAvg,
		FiveMinAvg: fiveMinAvg,
		FifteenMinAvg: fifteenMinAvg,
	}, nil
}