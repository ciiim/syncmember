package monitor

import (
	"log"
	"runtime"
	"time"
)

func MemoryUsageMB() float64 {
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	return float64(mem.Alloc) / 1024 / 1024
}

func ReportMemoryUsagePer(t time.Duration) {
	for {
		log.Printf("memory usage: %f MB", MemoryUsageMB())
		time.Sleep(t)
	}
}
