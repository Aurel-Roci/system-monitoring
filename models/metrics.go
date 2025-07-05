package models

import "time"

type ResultPtr struct {
	Type  string
	Data  interface{} // This can hold any type
	Error error
}

type LoadStats struct {
	OneMinAvg     float64
	FiveMinAvg    float64
	FifteenMinAvg float64
}

type MemoryStats struct {
	Available int64
	Free      int64
	Total     int64
	Timestamp time.Time
}

type CPUStats struct {
	UsagePct  float64
	CoreID    int
	Timestamp time.Time
}

type SystemTemp struct {
	TempInC   float64
	Timestamp time.Time
}
