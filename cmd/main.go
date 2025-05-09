package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type SystemVitals struct {
	CPUUsage    float64
	Memory      *mem.VirtualMemoryStat
	Disk        *disk.UsageStat
	Network     net.IOCountersStat
	HostInfo    *host.InfoStat
	Uptime      uint64
	LoadAvg     *load.AvgStat
	Processes   int
	Temperature []host.TemperatureStat
	GoRoutines  int
	GoMemAlloc  uint64
}

func collectSystemVitals() SystemVitals {
	vitals := SystemVitals{}

	// CPU Usage (1-second average)
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("CPU Usage: %v", err)
	} else if len(cpuPercents) > 0 {
		vitals.CPUUsage = cpuPercents[0]
	}

	// Memory Usage
	if memory, err := mem.VirtualMemory(); err != nil {
		log.Printf("Memory: %v", err)
	} else {
		vitals.Memory = memory
	}

	// Disk Usage (root partition)
	if diskUsage, err := disk.Usage("/"); err != nil {
		log.Printf("Disk: %v", err)
	} else {
		vitals.Disk = diskUsage
	}

	// Network I/O (sum all interfaces)
	if netIO, err := net.IOCounters(true); err != nil {
		log.Printf("Network: %v", err)
	} else {
		var total net.IOCountersStat
		for _, io := range netIO {
			total.BytesSent += io.BytesSent
			total.BytesRecv += io.BytesRecv
		}
		vitals.Network = total
	}

	// Host Information
	if hostInfo, err := host.Info(); err != nil {
		log.Printf("Host Info: %v", err)
	} else {
		vitals.HostInfo = hostInfo
	}

	// Uptime
	if uptime, err := host.Uptime(); err != nil {
		log.Printf("Uptime: %v", err)
	} else {
		vitals.Uptime = uptime
	}

	// Load Average
	if loadAvg, err := load.Avg(); err != nil {
		log.Printf("Load Average: %v", err)
	} else {
		vitals.LoadAvg = loadAvg
	}

	// Process Count
	if processes, err := process.Processes(); err != nil {
		log.Printf("Processes: %v", err)
	} else {
		vitals.Processes = len(processes)
	}

	// Temperature Sensors
	if temps, err := host.SensorsTemperatures(); err != nil {
		log.Printf("Temperature: %v", err)
	} else {
		vitals.Temperature = temps
	}

	// Go Runtime Metrics
	vitals.GoRoutines = runtime.NumGoroutine()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	vitals.GoMemAlloc = memStats.Alloc

	return vitals
}

func main() {
	vitals := collectSystemVitals()

	fmt.Println("=== System Vitals ===")
	fmt.Printf("CPU Usage: %.2f%%\n", vitals.CPUUsage)
	if vitals.Memory != nil {
		fmt.Printf("Memory: Total: %vB, Used: %vB (%.2f%%)\n",
			vitals.Memory.Total, vitals.Memory.Used, vitals.Memory.UsedPercent)
	}
	if vitals.Disk != nil {
		fmt.Printf("Disk: Total: %vB, Used: %vB (%.2f%%)\n",
			vitals.Disk.Total, vitals.Disk.Used, vitals.Disk.UsedPercent)
	}
	fmt.Printf("Network: Sent: %vB, Received: %vB\n",
		vitals.Network.BytesSent, vitals.Network.BytesRecv)
	if vitals.HostInfo != nil {
		fmt.Printf("Host: %s (%s %s), Uptime: %v\n",
			vitals.HostInfo.Hostname, vitals.HostInfo.Platform, vitals.HostInfo.PlatformVersion, vitals.Uptime)
	}
	if vitals.LoadAvg != nil {
		fmt.Printf("Load Avg: 1m=%.2f, 5m=%.2f, 15m=%.2f\n",
			vitals.LoadAvg.Load1, vitals.LoadAvg.Load5, vitals.LoadAvg.Load15)
	}
	fmt.Printf("Processes: %d\n", vitals.Processes)
	if len(vitals.Temperature) > 0 {
		fmt.Println("Temperatures:")
		for _, temp := range vitals.Temperature {
			fmt.Printf("  %s: %.2f°C\n", temp.SensorKey, temp.Temperature)
		}
	}
	fmt.Printf("Go Runtime: Goroutines=%d, Alloc=%vB\n",
		vitals.GoRoutines, vitals.GoMemAlloc)
}
