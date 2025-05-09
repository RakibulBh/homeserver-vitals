package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type SystemVitals struct {
	CPUUsage    float64                `json:"cpuUsage"`
	Memory      *mem.VirtualMemoryStat `json:"memory"`
	Disk        *disk.UsageStat        `json:"disk"`
	Network     net.IOCountersStat     `json:"network"`
	HostInfo    *host.InfoStat         `json:"hostInfo"`
	Uptime      uint64                 `json:"uptime"`
	LoadAvg     *load.AvgStat          `json:"loadAvg"`
	Processes   int                    `json:"processes"`
	Temperature []host.TemperatureStat `json:"temperature"`
	GoRoutines  int                    `json:"goRoutines"`
	GoMemAlloc  uint64                 `json:"goMemAlloc"`
}

func (app *application) initiateSSE(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check for flusher capability
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Register client disconnect detection
	notify := r.Context().Done()
	go func() {
		<-notify
		log.Println("Client disconnected")
	}()

	// Send SSE data at regular intervals
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Send initial data immediately
	sendVitalsData(w, flusher)

	// Keep sending data until client disconnects
	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			sendVitalsData(w, flusher)
		}
	}
}

func sendVitalsData(w http.ResponseWriter, flusher http.Flusher) {
	vitals := collectSystemVitals()

	jsonData, err := json.Marshal(vitals)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return
	}

	// Write the SSE data format
	_, err = fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if err != nil {
		log.Printf("Error writing to client: %v", err)
		return
	}

	// Ensure data is sent immediately
	flusher.Flush()
}

func collectSystemVitals() *SystemVitals {
	vitals := &SystemVitals{}

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

func (app *application) printVitals(w http.ResponseWriter, r *http.Request) {
	vitals := collectSystemVitals()

	fmt.Println("╒═══════════════════════════════╕")
	fmt.Println("│        SYSTEM VITALS         │")
	fmt.Println("╞═══════════════════════════════╡")

	// CPU
	fmt.Printf("│  \033[1mCPU\033[0m    %12.2f%%        │\n", vitals.CPUUsage)
	fmt.Println("├───────────────────────────────┤")

	// Memory
	if vitals.Memory != nil {
		fmt.Printf("│  \033[1mMEMORY\033[0m %15s       │\n", " ")
		fmt.Printf("│   Total: %-10v Used: %-6v │\n",
			vitals.Memory.Total, vitals.Memory.Used)
		fmt.Printf("│   Usage: %-10.2f%%%14s│\n",
			vitals.Memory.UsedPercent, " ")
		fmt.Println("├───────────────────────────────┤")
	}

	// Disk
	if vitals.Disk != nil {
		fmt.Printf("│  \033[1mDISK\033[0m  %15s       │\n", " ")
		fmt.Printf("│   Total: %-10v Used: %-6v │\n",
			vitals.Disk.Total, vitals.Disk.Used)
		fmt.Printf("│   Usage: %-10.2f%%%14s│\n",
			vitals.Disk.UsedPercent, " ")
		fmt.Println("├───────────────────────────────┤")
	}

	// Network
	fmt.Printf("│  \033[1mNETWORK\033[0m %13s       │\n", " ")
	fmt.Printf("│   ↑ %-10v  ↓ %-10v │\n",
		vitals.Network.BytesSent, vitals.Network.BytesRecv)
	fmt.Println("├───────────────────────────────┤")

	// Host Info
	if vitals.HostInfo != nil {
		fmt.Printf("│  \033[1mHOST\033[0m   %-23s │\n",
			vitals.HostInfo.Hostname)
		fmt.Printf("│   %s %-19s │\n",
			vitals.HostInfo.Platform, vitals.HostInfo.PlatformVersion)
		fmt.Printf("│   Uptime: %-19v │\n", time.Duration(vitals.Uptime)*time.Second)
		fmt.Println("├───────────────────────────────┤")
	}

	// Load & Processes
	if vitals.LoadAvg != nil {
		fmt.Printf("│  \033[1mLOAD\033[0m   1m:%-5.2f 5m:%-5.2f 15m:%-5.2f │\n",
			vitals.LoadAvg.Load1, vitals.LoadAvg.Load5, vitals.LoadAvg.Load15)
	}
	fmt.Printf("│  \033[1mPROCESSES\033[0m %19d │\n", vitals.Processes)
	fmt.Println("├───────────────────────────────┤")

	// Temperatures
	if len(vitals.Temperature) > 0 {
		fmt.Println("│  \033[1mTEMPERATURES\033[0m               │")
		for _, temp := range vitals.Temperature {
			fmt.Printf("│   %-20s %6.1f°C │\n",
				temp.SensorKey, temp.Temperature)
		}
		fmt.Println("├───────────────────────────────┤")
	}

	// Go Runtime
	fmt.Printf("│  \033[1mGO RUNTIME\033[0m                  │\n")
	fmt.Printf("│   Goroutines: %-15d │\n", vitals.GoRoutines)
	fmt.Printf("│   Memory: %-19v │\n", vitals.GoMemAlloc)
	fmt.Println("╘═══════════════════════════════╛")
}
