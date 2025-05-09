package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

// DiskInfo contains information about a disk/partition
type DiskInfo struct {
	MountPoint  string  `json:"mountPoint"`
	FileSystem  string  `json:"fileSystem"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

// NetworkInterface contains network interface information
type NetworkInterface struct {
	Name      string `json:"name"`
	IPAddress string `json:"ipAddress"`
	MacAddr   string `json:"macAddr"`
	BytesSent uint64 `json:"bytesSent"`
	BytesRecv uint64 `json:"bytesRecv"`
	IsUp      bool   `json:"isUp"`
}

// TopProcess contains information about top resource-consuming processes
type TopProcess struct {
	PID     int32   `json:"pid"`
	Name    string  `json:"name"`
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Command string  `json:"command"`
}

// HardwareInfo contains detailed hardware information
type HardwareInfo struct {
	CPUModel     string `json:"cpuModel"`
	CPUCores     int    `json:"cpuCores"`
	CPUThreads   int    `json:"cpuThreads"`
	TotalMemory  uint64 `json:"totalMemory"`
	SystemVendor string `json:"systemVendor"`
	SystemModel  string `json:"systemModel"`
}

// SystemVitals contains all system metrics
type SystemVitals struct {
	CPUUsage      float64                        `json:"cpuUsage"`
	CPUPerCore    []float64                      `json:"cpuPerCore"`
	Memory        *mem.VirtualMemoryStat         `json:"memory"`
	Swap          *mem.SwapMemoryStat            `json:"swap"`
	Disks         []DiskInfo                     `json:"disks"`
	Network       net.IOCountersStat             `json:"network"`
	NetworkIfaces []NetworkInterface             `json:"networkIfaces"`
	HostInfo      *host.InfoStat                 `json:"hostInfo"`
	Uptime        uint64                         `json:"uptime"`
	LoadAvg       *load.AvgStat                  `json:"loadAvg"`
	Processes     int                            `json:"processes"`
	Temperature   []host.TemperatureStat         `json:"temperature"`
	GoRoutines    int                            `json:"goRoutines"`
	GoMemAlloc    uint64                         `json:"goMemAlloc"`
	TopProcesses  []TopProcess                   `json:"topProcesses"`
	Hardware      HardwareInfo                   `json:"hardware"`
	LastUpdated   time.Time                      `json:"lastUpdated"`
	SystemUpdates int                            `json:"systemUpdates"`
	DiskIO        map[string]disk.IOCountersStat `json:"diskIO"`
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
	vitals := &SystemVitals{
		LastUpdated: time.Now(),
	}

	// CPU Usage (total and per core)
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("CPU Usage: %v", err)
	} else if len(cpuPercents) > 0 {
		vitals.CPUUsage = cpuPercents[0]
	}

	// CPU Usage per core
	perCore, err := cpu.Percent(time.Second, true)
	if err != nil {
		log.Printf("CPU Per Core: %v", err)
	} else {
		vitals.CPUPerCore = perCore
	}

	// Memory Usage
	if memory, err := mem.VirtualMemory(); err != nil {
		log.Printf("Memory: %v", err)
	} else {
		vitals.Memory = memory
	}

	// Swap Usage
	if swap, err := mem.SwapMemory(); err != nil {
		log.Printf("Swap: %v", err)
	} else {
		vitals.Swap = swap
	}

	// Disk Usage (all partitions)
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Printf("Disk Partitions: %v", err)
	} else {
		vitals.Disks = make([]DiskInfo, 0, len(partitions))
		for _, part := range partitions {
			usage, err := disk.Usage(part.Mountpoint)
			if err != nil {
				continue
			}

			diskInfo := DiskInfo{
				MountPoint:  part.Mountpoint,
				FileSystem:  part.Fstype,
				Total:       usage.Total,
				Used:        usage.Used,
				Free:        usage.Free,
				UsedPercent: usage.UsedPercent,
			}
			vitals.Disks = append(vitals.Disks, diskInfo)
		}
	}

	// Disk I/O stats
	diskIO, err := disk.IOCounters()
	if err != nil {
		log.Printf("Disk IO: %v", err)
	} else {
		vitals.DiskIO = diskIO
	}

	// Network I/O (sum all interfaces)
	if netIO, err := net.IOCounters(true); err != nil {
		log.Printf("Network: %v", err)
	} else {
		var total net.IOCountersStat

		// Collect network interfaces with IP addresses
		ifaces, _ := net.Interfaces()
		vitals.NetworkIfaces = make([]NetworkInterface, 0, len(ifaces))

		for _, io := range netIO {
			total.BytesSent += io.BytesSent
			total.BytesRecv += io.BytesRecv

			// Find matching interface to get IP
			for _, iface := range ifaces {
				if iface.Name == io.Name {
					netIface := NetworkInterface{
						Name:      io.Name,
						MacAddr:   iface.HardwareAddr,
						BytesSent: io.BytesSent,
						BytesRecv: io.BytesRecv,
						IsUp:      true, // Simplified
					}

					// Try to get IP address from interface name
					for _, addr := range ifaces {
						if addr.Name == iface.Name && len(addr.Addrs) > 0 {
							netIface.IPAddress = addr.Addrs[0].Addr
							break
						}
					}

					vitals.NetworkIfaces = append(vitals.NetworkIfaces, netIface)
					break
				}
			}
		}
		vitals.Network = total
	}

	// Host Information
	if hostInfo, err := host.Info(); err != nil {
		log.Printf("Host Info: %v", err)
	} else {
		vitals.HostInfo = hostInfo
	}

	// Hardware Info
	vitals.Hardware = collectHardwareInfo()

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

		// Get top processes by CPU and memory
		topProcesses := make([]TopProcess, 0, 5)
		for _, p := range processes {
			cpuPercent, _ := p.CPUPercent()
			memPercent, _ := p.MemoryPercent()
			name, _ := p.Name()
			cmdline, _ := p.Cmdline()

			// Only include processes with non-zero CPU usage
			if cpuPercent > 0 {
				topProc := TopProcess{
					PID:     p.Pid,
					Name:    name,
					CPU:     cpuPercent,
					Memory:  float64(memPercent),
					Command: cmdline,
				}

				topProcesses = append(topProcesses, topProc)
			}
		}

		// Sort by CPU usage (descending)
		for i := 0; i < len(topProcesses)-1; i++ {
			for j := i + 1; j < len(topProcesses); j++ {
				if topProcesses[i].CPU < topProcesses[j].CPU {
					topProcesses[i], topProcesses[j] = topProcesses[j], topProcesses[i]
				}
			}
		}

		// Keep only top 5
		if len(topProcesses) > 5 {
			topProcesses = topProcesses[:5]
		}

		vitals.TopProcesses = topProcesses
	}

	// Temperature Sensors
	if temps, err := host.SensorsTemperatures(); err != nil {
		log.Printf("Temperature: %v", err)
	} else {
		vitals.Temperature = temps
	}

	// System Updates Available
	vitals.SystemUpdates = checkForUpdates()

	// Go Runtime Metrics
	vitals.GoRoutines = runtime.NumGoroutine()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	vitals.GoMemAlloc = memStats.Alloc

	return vitals
}

// collectHardwareInfo gathers detailed hardware information
func collectHardwareInfo() HardwareInfo {
	info := HardwareInfo{}

	// CPU Info
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CPUModel = cpuInfo[0].ModelName
	}

	// CPU Cores/Threads
	counts, err := cpu.Counts(true)
	if err == nil {
		info.CPUThreads = counts
	}

	counts, err = cpu.Counts(false)
	if err == nil {
		info.CPUCores = counts
	}

	// Memory Total
	mem, err := mem.VirtualMemory()
	if err == nil {
		info.TotalMemory = mem.Total
	}

	// Try to get system vendor/model (Linux only)
	info.SystemVendor = getCommandOutput("cat /sys/devices/virtual/dmi/id/sys_vendor 2>/dev/null || echo 'Unknown'")
	info.SystemModel = getCommandOutput("cat /sys/devices/virtual/dmi/id/product_name 2>/dev/null || echo 'Unknown'")

	return info
}

// checkForUpdates counts available system updates
func checkForUpdates() int {
	updates := 0

	// Check for different package managers
	if runtime.GOOS == "linux" {
		// apt (Debian/Ubuntu)
		aptUpdates := getCommandOutput("apt list --upgradable 2>/dev/null | grep -v 'Listing...' | wc -l")
		if aptNum, err := parseCommandInt(aptUpdates); err == nil && aptNum > 0 {
			updates = aptNum
		}

		// yum/dnf (RHEL/CentOS/Fedora)
		if updates == 0 {
			yumUpdates := getCommandOutput("yum check-update --quiet | grep -v '^$' | wc -l")
			if yumNum, err := parseCommandInt(yumUpdates); err == nil {
				updates = yumNum
			}
		}
	} else if runtime.GOOS == "darwin" {
		// macOS (rough estimate using softwareupdate)
		macUpdates := getCommandOutput("softwareupdate -l 2>/dev/null | grep -i 'recommended' | wc -l")
		if macNum, err := parseCommandInt(macUpdates); err == nil {
			updates = macNum
		}
	}

	return updates
}

// getCommandOutput runs a shell command and returns its output
func getCommandOutput(cmdStr string) string {
	cmd := exec.Command("sh", "-c", cmdStr)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// parseCommandInt parses integer from command output
func parseCommandInt(output string) (int, error) {
	output = strings.TrimSpace(output)
	var value int
	_, err := fmt.Sscanf(output, "%d", &value)
	return value, err
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
	if vitals.Disks != nil {
		fmt.Printf("│  \033[1mDISKS\033[0m  %15s       │\n", " ")
		for _, disk := range vitals.Disks {
			fmt.Printf("│   %-10s %-10v Used: %-6v │\n",
				disk.MountPoint, disk.Total, disk.Used)
			fmt.Printf("│   Usage: %-10.2f%%%14s│\n",
				disk.UsedPercent, " ")
		}
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
