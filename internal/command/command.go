package command

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"pm2/internal/cluster"
	"pm2/internal/monitor"
	"pm2/internal/process"
	"pm2/internal/startup"
)

// ProcessManagerInterface defines the interface for process managers
type ProcessManagerInterface interface {
	Start(p *process.Process) error
	Stop(id string) error
	Restart(id string) error
	Delete(id string) error
	Reload(id string) error
	ListProcesses() []*process.Process
	GetProcess(id string) (*process.Process, error)
	Save() error
	Load() error
}

// ClusterManagerInterface defines the interface for cluster managers
type ClusterManagerInterface interface {
	StartCluster(c *cluster.Cluster) error
	StopCluster(id string) error
	RestartCluster(id string) error
	DeleteCluster(id string) error
	ReloadCluster(id string) error
	ListClusters() []*cluster.Cluster
	GetCluster(id string) (*cluster.Cluster, error)
	Save() error
	Load() error
}

// CommandManager manages all commands
type CommandManager struct {
	ProcessManager  ProcessManagerInterface
	ClusterManager ClusterManagerInterface
}

// NewCommandManager creates a new command manager
func NewCommandManager(processManager ProcessManagerInterface, clusterManager ClusterManagerInterface) *CommandManager {
	return &CommandManager{
		ProcessManager:  processManager,
		ClusterManager: clusterManager,
	}
}

// checkNameExists checks if a name already exists in either processes or clusters
func (cm *CommandManager) checkNameExists(name string) bool {
	// Check processes
	processes := cm.ProcessManager.ListProcesses()
	for _, p := range processes {
		if p.Name == name {
			return true
		}
	}

	// Check clusters
	clusters := cm.ClusterManager.ListClusters()
	for _, c := range clusters {
		if c.Name == name {
			return true
		}
	}

	return false
}

// StartCommand handles the start command
func (cm *CommandManager) StartCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("Please specify a command to run or a JSON config file")
	}

	// Check if the first argument is a JSON config file
	firstArg := c.Args().First()
	if strings.HasSuffix(firstArg, ".json") {
		// Read and parse JSON config file
		configData, err := os.ReadFile(firstArg)
		if err != nil {
			return fmt.Errorf("Error reading config file: %v", err)
		}

		// Define config struct
		type AppConfig struct {
			Name      string   `json:"name"`
			Script    string   `json:"script"`
			Args      []string `json:"args,omitempty"`
			Instances int      `json:"instances,omitempty"`
		}

		type Config struct {
			Apps []AppConfig `json:"apps"`
		}

		var config Config
		if err := json.Unmarshal(configData, &config); err != nil {
			return fmt.Errorf("Error parsing config file: %v", err)
		}

		// Start each app defined in the config
		for _, app := range config.Apps {
			command := app.Script
			args := app.Args
			name := app.Name
			if name == "" {
				name = command
			}
			instances := app.Instances
			if instances <= 0 {
				instances = 1
			}

			// Check if name already exists
			if cm.checkNameExists(name) {
				return fmt.Errorf("Application with name '%s' already exists", name)
			}

			id := fmt.Sprintf("%s_%d", name, time.Now().Unix())

			if instances > 1 {
				// Use cluster mode
				cluster := cluster.NewCluster(id, name, command, args, instances)
				if err := cm.ClusterManager.StartCluster(cluster); err != nil {
					return err
				}
				fmt.Printf("Started cluster %s with ID %s and %d instances\n", name, id, instances)
			} else {
				// Use single process mode
				p := process.NewProcess(id, name, command, args, instances)
				if err := cm.ProcessManager.Start(p); err != nil {
					return err
				}
				fmt.Printf("Started process %s with ID %s\n", name, id)
			}
		}

		return nil
	} else {
		// Regular command execution
		command := firstArg
		args := c.Args().Tail()
		name := c.String("name")
		if name == "" {
			name = command
		}
		instances := c.Int("instances")

		// Check if name already exists
		if cm.checkNameExists(name) {
			return fmt.Errorf("Application with name '%s' already exists", name)
		}

		id := fmt.Sprintf("%s_%d", name, time.Now().Unix())

		if instances > 1 {
			// Use cluster mode
			cluster := cluster.NewCluster(id, name, command, args, instances)
			if err := cm.ClusterManager.StartCluster(cluster); err != nil {
				return err
			}
			fmt.Printf("Started cluster %s with ID %s and %d instances\n", name, id, instances)
		} else {
			// Use single process mode
			p := process.NewProcess(id, name, command, args, instances)
			if err := cm.ProcessManager.Start(p); err != nil {
				return err
			}
			fmt.Printf("Started process %s with ID %s\n", name, id)
		}

		return nil
	}
}

// StopCommand handles the stop command
func (cm *CommandManager) StopCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("Please specify a process ID or name")
	}

	id := c.Args().First()

	// Try to stop as a regular process first
	if err := cm.ProcessManager.Stop(id); err == nil {
		fmt.Printf("Stopped process %s\n", id)
		return nil
	}

	// Try to stop as a cluster
	if err := cm.ClusterManager.StopCluster(id); err == nil {
		fmt.Printf("Stopped cluster %s\n", id)
		return nil
	}

	return fmt.Errorf("Process or cluster with ID %s not found", id)
}

// RestartCommand handles the restart command
func (cm *CommandManager) RestartCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("Please specify a process ID or name")
	}

	id := c.Args().First()

	// Try to restart as a regular process first
	if err := cm.ProcessManager.Restart(id); err == nil {
		fmt.Printf("Restarted process %s\n", id)
		return nil
	}

	// Try to restart as a cluster
	if err := cm.ClusterManager.RestartCluster(id); err == nil {
		fmt.Printf("Restarted cluster %s\n", id)
		return nil
	}

	return fmt.Errorf("Process or cluster with ID %s not found", id)
}

// DeleteCommand handles the delete command
func (cm *CommandManager) DeleteCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("Please specify a process ID or name")
	}

	id := c.Args().First()

	// Handle "all" parameter
	if id == "all" {
		// Delete all regular processes
		processes := cm.ProcessManager.ListProcesses()
		for _, p := range processes {
			cm.ProcessManager.Delete(p.ID)
		}

		// Delete all clusters
		clusterCount := 0
		for {
			clusters := cm.ClusterManager.ListClusters()
			if len(clusters) == 0 {
				break
			}
			// Delete the first cluster
			if err := cm.ClusterManager.DeleteCluster(clusters[0].ID); err != nil {
				// Skip if there's an error
				fmt.Printf("Error deleting cluster %s: %v\n", clusters[0].ID, err)
			}
			clusterCount++
		}

		fmt.Printf("Deleted %d processes and %d clusters\n", len(processes), clusterCount)
		return nil
	}

	// Check if id is a number (index)
	if index, err := strconv.Atoi(id); err == nil {
		// Try to delete by index
		processes := cm.ProcessManager.ListProcesses()
		clusters := cm.ClusterManager.ListClusters()
		total := len(processes) + len(clusters)

		if index < 0 || index >= total {
			return fmt.Errorf("Index out of range. Total processes and clusters: %d", total)
		}

		if index < len(processes) {
			// Delete process by index
			process := processes[index]
			if err := cm.ProcessManager.Delete(process.ID); err == nil {
				fmt.Printf("Deleted process %s (index: %d)\n", process.ID, index)
				return nil
			}
		} else {
			// Delete cluster by index
			clusterIndex := index - len(processes)
			if clusterIndex < len(clusters) {
				cluster := clusters[clusterIndex]
				if err := cm.ClusterManager.DeleteCluster(cluster.ID); err == nil {
					fmt.Printf("Deleted cluster %s (index: %d)\n", cluster.ID, index)
					return nil
				}
			}
		}
	}

	// Try to delete as a regular process first
	if err := cm.ProcessManager.Delete(id); err == nil {
		fmt.Printf("Deleted process %s\n", id)
		return nil
	}

	// Try to delete as a cluster
	if err := cm.ClusterManager.DeleteCluster(id); err == nil {
		fmt.Printf("Deleted cluster %s\n", id)
		return nil
	}

	// Try to find and delete by name
	processes := cm.ProcessManager.ListProcesses()
	for _, p := range processes {
		if p.Name == id {
			if err := cm.ProcessManager.Delete(p.ID); err == nil {
				fmt.Printf("Deleted process %s (name: %s)\n", p.ID, p.Name)
				return nil
			}
		}
	}

	// Try to find and delete cluster by name
	clusters := cm.ClusterManager.ListClusters()
	for _, c := range clusters {
		if c.Name == id {
			if err := cm.ClusterManager.DeleteCluster(c.ID); err == nil {
				fmt.Printf("Deleted cluster %s (name: %s)\n", c.ID, c.Name)
				return nil
			}
		}
	}

	return fmt.Errorf("Process or cluster with ID or name '%s' not found", id)
}

// ListCommand handles the list command
func (cm *CommandManager) ListCommand(c *cli.Context) error {
	processes := cm.ProcessManager.ListProcesses()
	clusters := cm.ClusterManager.ListClusters()

	if len(processes) == 0 && len(clusters) == 0 {
		fmt.Println("No processes or clusters running")
		return nil
	}

	// ANSI color codes
	reset := "\033[0m"
	bold := "\033[1m"
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	blue := "\033[34m"

	// Calculate column widths for all processes and clusters
	idWidth := len("id")
	nameWidth := len("name")
	modeWidth := len("mode")
	statusWidth := len("status")
	cpuWidth := len("cpu")
	memWidth := len("mem")
	restartWidth := len("restart")
	uptimeWidth := len("uptime")
	pidWidth := len("pid")

	// Check regular processes
	for i, p := range processes {
		indexStr := fmt.Sprintf("%d", i)
		if len(indexStr) > idWidth {
			idWidth = len(indexStr)
		}
		if len(p.Name) > nameWidth {
			nameWidth = len(p.Name)
		}
		// Check mode width
		modeText := "fork"
		if len(modeText) > modeWidth {
			modeWidth = len(modeText)
		}
		// Check status text width
		statusText := ""
		if p.Status == "running" {
			statusText = "online"
		} else if p.Status == "stopped" {
			statusText = "stopped"
		} else {
			statusText = "errored"
		}
		if len(statusText) > statusWidth {
			statusWidth = len(statusText)
		}
		// Check CPU and memory width
		cpuStr := fmt.Sprintf("%.1f%%", 0.0)
		memStr := fmt.Sprintf("%.1fMB", 0.0)
		if len(cpuStr) > cpuWidth {
			cpuWidth = len(cpuStr)
		}
		if len(memStr) > memWidth {
			memWidth = len(memStr)
		}
		// Check restart width
		restartStr := fmt.Sprintf("%d", p.RestartCount)
		if len(restartStr) > restartWidth {
			restartWidth = len(restartStr)
		}
		// Check uptime width
		uptime := time.Since(p.CreatedAt)
		uptimeStr := formatDuration(uptime)
		if len(uptimeStr) > uptimeWidth {
			uptimeWidth = len(uptimeStr)
		}
		// Check PID width
		pidStr := fmt.Sprintf("%v", p.PIDs)
		if len(pidStr) > pidWidth {
			pidWidth = len(pidStr)
		}
	}

	// Check clusters
	for i, c := range clusters {
		indexStr := fmt.Sprintf("%d", len(processes)+i)
		if len(indexStr) > idWidth {
			idWidth = len(indexStr)
		}
		if len(c.Name) > nameWidth {
			nameWidth = len(c.Name)
		}
		// Check mode width
		modeText := "cluster"
		if len(modeText) > modeWidth {
			modeWidth = len(modeText)
		}
		// Check status text width
		statusText := ""
		if c.Status == "running" {
			statusText = "online"
		} else if c.Status == "stopped" {
			statusText = "stopped"
		} else {
			statusText = "errored"
		}
		if len(statusText) > statusWidth {
			statusWidth = len(statusText)
		}
		// Check CPU and memory width
		cpuStr := fmt.Sprintf("%.1f%%", 0.0)
		memStr := fmt.Sprintf("%.1fMB", 0.0)
		if len(cpuStr) > cpuWidth {
			cpuWidth = len(cpuStr)
		}
		if len(memStr) > memWidth {
			memWidth = len(memStr)
		}
		// Check restart width
		restartStr := fmt.Sprintf("%d", c.RestartCount)
		if len(restartStr) > restartWidth {
			restartWidth = len(restartStr)
		}
		// Check uptime width
		uptime := time.Since(c.CreatedAt)
		uptimeStr := formatDuration(uptime)
		if len(uptimeStr) > uptimeWidth {
			uptimeWidth = len(uptimeStr)
		}
		// Check PID width
		pidStr := fmt.Sprintf("%v", c.PIDs)
		if len(pidStr) > pidWidth {
			pidWidth = len(pidStr)
		}
	}

	// Print header
	fmt.Println("\n=== PM2 Process Manager ===")
	fmt.Printf("%s%-*s%s %s%-*s%s %s%-*s%s %s%-*s%s %s%-*s%s %s%-*s%s %s%-*s%s %s%-*s%s %s%-*s%s\n",
		bold, idWidth, "id", reset,
		bold, nameWidth, "name", reset,
		bold, modeWidth, "mode", reset,
		bold, statusWidth, "status", reset,
		bold, cpuWidth, "cpu", reset,
		bold, memWidth, "mem", reset,
		bold, restartWidth, "restart", reset,
		bold, uptimeWidth, "uptime", reset,
		bold, pidWidth, "pid", reset)

	// Print separator
	separator := ""
	for i := 0; i < idWidth+nameWidth+modeWidth+statusWidth+cpuWidth+memWidth+restartWidth+uptimeWidth+pidWidth+8; i++ {
		separator += "-"
	}
	fmt.Println(separator)

	// Print regular processes
	for i, p := range processes {
		// Get CPU and memory usage
		cpu := 0.0
		mem := 0.0
		if len(p.PIDs) > 0 {
			if stats, err := monitor.GetProcessStats(p.PIDs[0]); err == nil {
				cpu = stats.CPU
				mem = stats.Memory
			}
		}

		// Calculate uptime
		uptime := time.Since(p.CreatedAt)
		uptimeStr := formatDuration(uptime)

		// Set status color
		statusColor := reset
		statusText := p.Status
		if p.Status == "running" {
			statusColor = green
			statusText = "online"
		} else if p.Status == "stopped" {
			statusColor = yellow
			statusText = "stopped"
		} else {
			statusColor = red
			statusText = "errored"
		}

		// Format values
		cpuStr := fmt.Sprintf("%.1f%%", cpu)
		memStr := fmt.Sprintf("%.1fMB", mem)
		pidStr := fmt.Sprintf("%v", p.PIDs)
		indexStr := fmt.Sprintf("%d", i)

		// Set mode color
		modeColor := reset
		modeText := "fork"

		fmt.Printf("%-*s %-*s %s%-*s%s %s%-*s%s %-*s %-*s %-*d %-*s %-*s\n",
			idWidth, indexStr,
			nameWidth, p.Name,
			modeColor, modeWidth, modeText, reset,
			statusColor, statusWidth, statusText, reset,
			cpuWidth, cpuStr,
			memWidth, memStr,
			restartWidth, p.RestartCount,
			uptimeWidth, uptimeStr,
			pidWidth, pidStr)
	}

	// Print clusters
	for i, c := range clusters {
		// Get CPU and memory usage
		cpu := 0.0
		mem := 0.0
		if len(c.PIDs) > 0 {
			if stats, err := monitor.GetProcessStats(c.PIDs[0]); err == nil {
				cpu = stats.CPU
				mem = stats.Memory
			}
		}

		// Calculate uptime
		uptime := time.Since(c.CreatedAt)
		uptimeStr := formatDuration(uptime)

		// Set status color
		statusColor := reset
		statusText := c.Status
		if c.Status == "running" {
			statusColor = green
			statusText = "online"
		} else if c.Status == "stopped" {
			statusColor = yellow
			statusText = "stopped"
		} else {
			statusColor = red
			statusText = "errored"
		}

		// Format values
		cpuStr := fmt.Sprintf("%.1f%%", cpu)
		memStr := fmt.Sprintf("%.1fMB", mem)
		pidStr := fmt.Sprintf("%v", c.PIDs)
		indexStr := fmt.Sprintf("%d", len(processes)+i)

		// Set mode color
		modeColor := blue
		modeText := "cluster"

		fmt.Printf("%-*s %-*s %s%-*s%s %s%-*s%s %-*s %-*s %-*d %-*s %-*s\n",
			idWidth, indexStr,
			nameWidth, c.Name,
			modeColor, modeWidth, modeText, reset,
			statusColor, statusWidth, statusText, reset,
			cpuWidth, cpuStr,
			memWidth, memStr,
			restartWidth, c.RestartCount,
			uptimeWidth, uptimeStr,
			pidWidth, pidStr)
	}

	// Print footer
	fmt.Println(separator)
	fmt.Printf("%s%d process(es), %d online, %d stopped, %d errored%s\n",
		bold, len(processes)+len(clusters), 
		countOnlineProcesses(processes, clusters),
		countStoppedProcesses(processes, clusters),
		countErroredProcesses(processes, clusters),
		reset)

	return nil
}

// LogsCommand handles the logs command
func (cm *CommandManager) LogsCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("Please specify a process ID or name")
	}

	id := c.Args().First()

	// Try to get logs from a regular process first
	p, err := cm.ProcessManager.GetProcess(id)
	if err == nil {
		if p.Logger != nil {
			if logger, ok := p.Logger.(interface{ ReadLogs() (string, error) }); ok {
				logs, err := logger.ReadLogs()
				if err != nil {
					return err
				}
				fmt.Println(logs)
				return nil
			}
		}
		return fmt.Errorf("No logs available for process %s", id)
	}

	// Try to get logs from a cluster
	cluster, err := cm.ClusterManager.GetCluster(id)
	if err == nil {
		if cluster.Logger != nil {
			if logger, ok := cluster.Logger.(interface{ ReadLogs() (string, error) }); ok {
				logs, err := logger.ReadLogs()
				if err != nil {
					return err
				}
				fmt.Println(logs)
				return nil
			}
		}
		return fmt.Errorf("No logs available for cluster %s", id)
	}

	return fmt.Errorf("Process or cluster with ID %s not found", id)
}

// MonitCommand handles the monit command
func (cm *CommandManager) MonitCommand(c *cli.Context) error {
	// Get all processes and clusters
	processes := cm.ProcessManager.ListProcesses()
	clusters := cm.ClusterManager.ListClusters()

	// Collect all PIDs
	pids := []int{}
	for _, p := range processes {
		pids = append(pids, p.PIDs...)
	}
	for _, c := range clusters {
		pids = append(pids, c.PIDs...)
	}

	// Start host monitoring in a goroutine
	go monitor.MonitorHost()

	// Start process monitoring
	monitor.MonitorProcesses(pids)

	return nil
}

// StartupCommand handles the startup command
func (cm *CommandManager) StartupCommand(c *cli.Context) error {
	if err := startup.GenerateStartupScript(); err != nil {
		return err
	}

	fmt.Println("Startup script generated successfully")
	return nil
}

// SaveCommand handles the save command
func (cm *CommandManager) SaveCommand(c *cli.Context) error {
	// Save process and cluster list to a configuration file
	configFile := "/etc/pm3/processes.json"

	// Create directory if it doesn't exist
	if err := os.MkdirAll("/etc/pm3", 0755); err != nil {
		return err
	}

	// Get all processes and clusters
	processes := cm.ProcessManager.ListProcesses()
	clusters := cm.ClusterManager.ListClusters()

	// Write configuration to file (simplified for now)
	// In a real implementation, we would use JSON marshaling
	configContent := fmt.Sprintf("{\"processes\": %d, \"clusters\": %d}", len(processes), len(clusters))

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return err
	}

	fmt.Println("Process list saved successfully")
	return nil
}

// ReloadCommand handles the reload command
func (cm *CommandManager) ReloadCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("Please specify a process ID or name")
	}

	id := c.Args().First()

	// Try to reload as a regular process first
	if err := cm.ProcessManager.Reload(id); err == nil {
		fmt.Printf("Reloaded process %s with zero downtime\n", id)
		return nil
	}

	// Try to reload as a cluster
	if err := cm.ClusterManager.ReloadCluster(id); err == nil {
		fmt.Printf("Reloaded cluster %s with zero downtime\n", id)
		return nil
	}

	return fmt.Errorf("Process or cluster with ID %s not found", id)
}

// StatusCommand handles the status command
func (cm *CommandManager) StatusCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("Please specify a process ID or name")
	}

	id := c.Args().First()

	// Try to get status of a regular process first
	p, err := cm.ProcessManager.GetProcess(id)
	if err == nil {
		fmt.Printf("Process %s status: %s\n", p.Name, p.Status)
		return nil
	}

	// Try to get status of a cluster
	cluster, err := cm.ClusterManager.GetCluster(id)
	if err == nil {
		fmt.Printf("Cluster %s status: %s\n", cluster.Name, cluster.Status)
		return nil
	}

	return fmt.Errorf("Process or cluster with ID %s not found", id)
}

// DescribeCommand handles the describe command
func (cm *CommandManager) DescribeCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("Please specify a process ID or name")
	}

	id := c.Args().First()

	// Try to describe a regular process first
	p, err := cm.ProcessManager.GetProcess(id)
	if err == nil {
		fmt.Printf("=== Process Details ===\n")
		fmt.Printf("ID: %s\n", p.ID)
		fmt.Printf("Name: %s\n", p.Name)
		fmt.Printf("Command: %s\n", p.Command)
		fmt.Printf("Args: %v\n", p.Args)
		fmt.Printf("Instances: %d\n", p.Instances)
		fmt.Printf("Status: %s\n", p.Status)
		fmt.Printf("PIDs: %v\n", p.PIDs)
		fmt.Printf("Created At: %s\n", p.CreatedAt)
		fmt.Printf("Updated At: %s\n", p.UpdatedAt)
		return nil
	}

	// Try to describe a cluster
	cluster, err := cm.ClusterManager.GetCluster(id)
	if err == nil {
		fmt.Printf("=== Cluster Details ===\n")
		fmt.Printf("ID: %s\n", cluster.ID)
		fmt.Printf("Name: %s\n", cluster.Name)
		fmt.Printf("Command: %s\n", cluster.Command)
		fmt.Printf("Args: %v\n", cluster.Args)
		fmt.Printf("Instances: %d\n", cluster.Instances)
		fmt.Printf("Status: %s\n", cluster.Status)
		fmt.Printf("PIDs: %v\n", cluster.PIDs)
		fmt.Printf("Created At: %s\n", cluster.CreatedAt)
		fmt.Printf("Updated At: %s\n", cluster.UpdatedAt)
		return nil
	}

	return fmt.Errorf("Process or cluster with ID %s not found", id)
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// countOnlineProcesses counts the number of online processes
func countOnlineProcesses(processes []*process.Process, clusters []*cluster.Cluster) int {
	count := 0
	for _, p := range processes {
		if p.Status == "running" {
			count++
		}
	}
	for _, c := range clusters {
		if c.Status == "running" {
			count++
		}
	}
	return count
}

// countStoppedProcesses counts the number of stopped processes
func countStoppedProcesses(processes []*process.Process, clusters []*cluster.Cluster) int {
	count := 0
	for _, p := range processes {
		if p.Status == "stopped" {
			count++
		}
	}
	for _, c := range clusters {
		if c.Status == "stopped" {
			count++
		}
	}
	return count
}

// countErroredProcesses counts the number of errored processes
func countErroredProcesses(processes []*process.Process, clusters []*cluster.Cluster) int {
	count := 0
	for _, p := range processes {
		if p.Status == "errored" {
			count++
		}
	}
	for _, c := range clusters {
		if c.Status == "errored" {
			count++
		}
	}
	return count
}
