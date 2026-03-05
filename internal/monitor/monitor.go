package monitor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ProcessStats represents process statistics
type ProcessStats struct {
	PID        int       `json:"pid"`
	Name       string    `json:"name"`
	CPU        float64   `json:"cpu"`
	Memory     float64   `json:"memory"`
	Uptime     time.Duration `json:"uptime"`
	Status     string    `json:"status"`
}

// HostStats represents host statistics
type HostStats struct {
	CPU        float64   `json:"cpu"`
	Memory     float64   `json:"memory"`
	Disk       float64   `json:"disk"`
	Network    float64   `json:"network"`
	LoadAvg    []float64 `json:"load_avg"`
	Uptime     time.Duration `json:"uptime"`
	OS         string    `json:"os"`
	Arch       string    `json:"arch"`
}

// GetProcessStats returns statistics for a process
func GetProcessStats(pid int) (*ProcessStats, error) {
	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	// Get process status
	status := "unknown"
	if process != nil {
		status = "running"
	}

	// Get process uptime (simplified)
	uptime := time.Since(time.Now().Add(-time.Hour))

	// Get CPU and memory usage for Linux
	cpu := 0.0
	mem := 0.0

	if runtime.GOOS == "linux" {
		// Read /proc/{pid}/stat for CPU usage
		statPath := fmt.Sprintf("/proc/%d/stat", pid)
		statData, err := os.ReadFile(statPath)
		if err == nil {
			parts := strings.Fields(string(statData))
			if len(parts) >= 14 {
				// Calculate CPU usage (simplified)
				// For a real implementation, we would need to track previous values
				cpu = 0.1 // Placeholder for actual calculation
			}
		}

		// Read /proc/{pid}/status for memory usage
		statusPath := fmt.Sprintf("/proc/%d/status", pid)
		statusData, err := os.ReadFile(statusPath)
		if err == nil {
			lines := strings.Split(string(statusData), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "VmRSS:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						if val, err := strconv.ParseFloat(parts[1], 64); err == nil {
							mem = val / 1024 // Convert to MB
						}
					}
					break
				}
			}
		}
	}

	return &ProcessStats{
		PID:    pid,
		Name:   fmt.Sprintf("Process %d", pid),
		CPU:    cpu,
		Memory: mem,
		Uptime: uptime,
		Status: status,
	}, nil
}

// GetHostStats returns statistics for the host
func GetHostStats() (*HostStats, error) {
	// Get OS and architecture
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Get load average
	loadAvg := []float64{0.0, 0.0, 0.0}
	if os == "linux" {
		output, err := exec.Command("cat", "/proc/loadavg").Output()
		if err == nil {
			parts := strings.Fields(string(output))
			if len(parts) >= 3 {
				for i := 0; i < 3; i++ {
					if val, err := strconv.ParseFloat(parts[i], 64); err == nil {
						loadAvg[i] = val
					}
				}
			}
		}
	}

	// Get uptime
	uptime := time.Since(time.Now().Add(-24 * time.Hour)) // Placeholder

	// Get CPU usage (simplified)
	cpu := 0.0 // Placeholder

	// Get memory usage (simplified)
	memory := 0.0 // Placeholder

	// Get disk usage (simplified)
	disk := 0.0 // Placeholder

	// Get network usage (simplified)
	network := 0.0 // Placeholder

	return &HostStats{
		CPU:     cpu,
		Memory:  memory,
		Disk:    disk,
		Network: network,
		LoadAvg: loadAvg,
		Uptime:  uptime,
		OS:      os,
		Arch:    arch,
	}, nil
}

// MonitorProcesses monitors processes and prints statistics
func MonitorProcesses(pids []int) {
	for {
		fmt.Println("=== Process Monitoring ===")
		for _, pid := range pids {
			stats, err := GetProcessStats(pid)
			if err == nil {
				fmt.Printf("PID: %d, Name: %s, Status: %s, Uptime: %s\n", 
					stats.PID, stats.Name, stats.Status, stats.Uptime)
			}
		}
		fmt.Println()
		time.Sleep(2 * time.Second)
	}
}

// MonitorHost monitors the host and prints statistics
func MonitorHost() {
	for {
		fmt.Println("=== Host Monitoring ===")
		stats, err := GetHostStats()
		if err == nil {
			fmt.Printf("OS: %s, Arch: %s, Load Avg: %v, Uptime: %s\n", 
				stats.OS, stats.Arch, stats.LoadAvg, stats.Uptime)
		}
		fmt.Println()
		time.Sleep(2 * time.Second)
	}
}
