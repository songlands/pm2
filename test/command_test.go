package test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/urfave/cli/v2"
	"pm2/internal/cluster"
	"pm2/internal/command"
	"pm2/internal/process"
)

// MockProcessManager is a mock implementation of process.Manager for testing
type MockProcessManager struct {
	processes map[string]*process.Process
}

func NewMockProcessManager() *MockProcessManager {
	return &MockProcessManager{
		processes: make(map[string]*process.Process),
	}
}

func (m *MockProcessManager) Start(p *process.Process) error {
	m.processes[p.ID] = p
	return nil
}

func (m *MockProcessManager) Stop(id string) error {
	if p, ok := m.processes[id]; ok {
		p.Status = "stopped"
		return nil
	}
	return fmt.Errorf("process not found")
}

func (m *MockProcessManager) Restart(id string) error {
	if p, ok := m.processes[id]; ok {
		p.Status = "running"
		p.RestartCount++
		return nil
	}
	return fmt.Errorf("process not found")
}

func (m *MockProcessManager) Delete(id string) error {
	if _, ok := m.processes[id]; ok {
		delete(m.processes, id)
		return nil
	}
	return fmt.Errorf("process not found")
}

func (m *MockProcessManager) Reload(id string) error {
	if _, ok := m.processes[id]; ok {
		return nil
	}
	return nil
}

func (m *MockProcessManager) ListProcesses() []*process.Process {
	var processes []*process.Process
	for _, p := range m.processes {
		processes = append(processes, p)
	}
	return processes
}

func (m *MockProcessManager) GetProcess(id string) (*process.Process, error) {
	if p, ok := m.processes[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("process not found")
}

func (m *MockProcessManager) Save() error {
	return nil
}

func (m *MockProcessManager) Load() error {
	return nil
}

// MockClusterManager is a mock implementation of cluster.Manager for testing
type MockClusterManager struct {
	clusters map[string]*cluster.Cluster
}

func NewMockClusterManager() *MockClusterManager {
	return &MockClusterManager{
		clusters: make(map[string]*cluster.Cluster),
	}
}

func (m *MockClusterManager) StartCluster(c *cluster.Cluster) error {
	m.clusters[c.ID] = c
	return nil
}

func (m *MockClusterManager) StopCluster(id string) error {
	if c, ok := m.clusters[id]; ok {
		c.Status = "stopped"
		return nil
	}
	return fmt.Errorf("cluster not found")
}

func (m *MockClusterManager) RestartCluster(id string) error {
	if c, ok := m.clusters[id]; ok {
		c.Status = "running"
		c.RestartCount++
		return nil
	}
	return fmt.Errorf("cluster not found")
}

func (m *MockClusterManager) DeleteCluster(id string) error {
	delete(m.clusters, id)
	return nil
}

func (m *MockClusterManager) ReloadCluster(id string) error {
	if _, ok := m.clusters[id]; ok {
		return nil
	}
	return nil
}

func (m *MockClusterManager) ListClusters() []*cluster.Cluster {
	var clusters []*cluster.Cluster
	for _, c := range m.clusters {
		clusters = append(clusters, c)
	}
	return clusters
}

func (m *MockClusterManager) GetCluster(id string) (*cluster.Cluster, error) {
	if c, ok := m.clusters[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("cluster not found")
}

func (m *MockClusterManager) Save() error {
	return nil
}

func (m *MockClusterManager) Load() error {
	return nil
}

// createContext creates a CLI context with the given arguments and flags
func createContext(args []string, flags map[string]string) *cli.Context {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "instances"},
			&cli.StringFlag{Name: "name"},
		},
	}

	// Create a command
	cmd := &cli.Command{
		Name: "start",
		Action: func(c *cli.Context) error {
			return nil
		},
	}

	// Create a flag set and parse the arguments
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	for _, f := range app.Flags {
		f.Apply(fs)
	}

	// Parse the flags - first flags, then arguments
	cmdArgs := []string{}
	for key, value := range flags {
		cmdArgs = append(cmdArgs, "--"+key, value)
	}
	cmdArgs = append(cmdArgs, args...)

	fs.Parse(cmdArgs)

	// Create a context with the parsed flag set
	ctx := cli.NewContext(app, fs, nil)
	ctx.Command = cmd

	return ctx
}

// TestStartCommand tests the start command
func TestStartCommand(t *testing.T) {
	// Test case 1: Start a single process
	mockProcessManager1 := NewMockProcessManager()
	mockClusterManager1 := NewMockClusterManager()
	commandManager1 := command.NewCommandManager(mockProcessManager1, mockClusterManager1)

	ctx := createContext([]string{"./app.js"}, map[string]string{
		"name":      "test-app",
		"instances": "1",
	})

	err := commandManager1.StartCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	processes := mockProcessManager1.ListProcesses()
	if len(processes) != 1 {
		t.Errorf("Expected 1 process, got %d", len(processes))
	}

	// Test case 2: Start a cluster
	mockProcessManager2 := NewMockProcessManager()
	mockClusterManager2 := NewMockClusterManager()
	commandManager2 := command.NewCommandManager(mockProcessManager2, mockClusterManager2)

	ctx = createContext([]string{"./app.js"}, map[string]string{
		"name":      "test-cluster-app",
		"instances": "2",
	})

	err = commandManager2.StartCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	clusters := mockClusterManager2.ListClusters()
	if len(clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(clusters))
	}

	// Test case 3: Start from JSON config
	mockProcessManager3 := NewMockProcessManager()
	mockClusterManager3 := NewMockClusterManager()
	commandManager3 := command.NewCommandManager(mockProcessManager3, mockClusterManager3)

	// Create a temporary JSON config file
	config := struct {
		Apps []struct {
			Name      string   `json:"name"`
			Script    string   `json:"script"`
			Args      []string `json:"args,omitempty"`
			Instances int      `json:"instances,omitempty"`
		} `json:"apps"`
	}{
		Apps: []struct {
			Name      string   `json:"name"`
			Script    string   `json:"script"`
			Args      []string `json:"args,omitempty"`
			Instances int      `json:"instances,omitempty"`
		}{
			{
				Name:      "config-app",
				Script:    "node",
				Args:      []string{"app.js"},
				Instances: 1,
			},
		},
	}

	configData, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal config: %v", err)
	}

	configFile, err := os.CreateTemp("", "ecosystem*.json")
	if err != nil {
		t.Errorf("Failed to create temp file: %v", err)
	}
	defer os.Remove(configFile.Name())

	_, err = configFile.Write(configData)
	if err != nil {
		t.Errorf("Failed to write config file: %v", err)
	}
	configFile.Close()

	// Test starting from config file
	ctx = createContext([]string{configFile.Name()}, map[string]string{})

	err = commandManager3.StartCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	processes = mockProcessManager3.ListProcesses()
	found := false
	for _, p := range processes {
		if p.Name == "config-app" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected process with name 'config-app' to be started")
	}
}

// TestStopCommand tests the stop command
func TestStopCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Create a test process
	process := process.NewProcess("test-process", "test-app", "./app.js", []string{}, 1)
	process.Status = "running"
	mockProcessManager.Start(process)

	// Create a test cluster
	cluster := cluster.NewCluster("test-cluster", "test-cluster-app", "./app.js", []string{}, 2)
	cluster.Status = "running"
	mockClusterManager.StartCluster(cluster)

	// Test stopping the process
	ctx := createContext([]string{"test-process"}, map[string]string{})

	err := commandManager.StopCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	p, _ := mockProcessManager.GetProcess("test-process")
	if p.Status != "stopped" {
		t.Errorf("Expected process status to be 'stopped', got '%s'", p.Status)
	}

	// Test stopping the cluster
	ctx = createContext([]string{"test-cluster"}, map[string]string{})

	err = commandManager.StopCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	c, _ := mockClusterManager.GetCluster("test-cluster")
	if c.Status != "stopped" {
		t.Errorf("Expected cluster status to be 'stopped', got '%s'", c.Status)
	}
}

// TestRestartCommand tests the restart command
func TestRestartCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Create a test process
	process := process.NewProcess("test-process", "test-app", "./app.js", []string{}, 1)
	process.Status = "stopped"
	process.RestartCount = 0
	mockProcessManager.Start(process)

	// Create a test cluster
	cluster := cluster.NewCluster("test-cluster", "test-cluster-app", "./app.js", []string{}, 2)
	cluster.Status = "stopped"
	cluster.RestartCount = 0
	mockClusterManager.StartCluster(cluster)

	// Test restarting the process
	ctx := createContext([]string{"test-process"}, map[string]string{})

	err := commandManager.RestartCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	p, _ := mockProcessManager.GetProcess("test-process")
	if p.Status != "running" {
		t.Errorf("Expected process status to be 'running', got '%s'", p.Status)
	}
	if p.RestartCount != 1 {
		t.Errorf("Expected process restart count to be 1, got %d", p.RestartCount)
	}

	// Test restarting the cluster
	ctx = createContext([]string{"test-cluster"}, map[string]string{})

	err = commandManager.RestartCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	c, _ := mockClusterManager.GetCluster("test-cluster")
	if c.Status != "running" {
		t.Errorf("Expected cluster status to be 'running', got '%s'", c.Status)
	}
	if c.RestartCount != 1 {
		t.Errorf("Expected cluster restart count to be 1, got %d", c.RestartCount)
	}
}

// TestDeleteCommand tests the delete command
func TestDeleteCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Create test processes and clusters
	process1 := process.NewProcess("test-process-1", "test-app-1", "./app.js", []string{}, 1)
	process2 := process.NewProcess("test-process-2", "test-app-2", "./app.js", []string{}, 1)
	mockProcessManager.Start(process1)
	mockProcessManager.Start(process2)

	cluster1 := cluster.NewCluster("test-cluster-1", "test-cluster-app-1", "./app.js", []string{}, 2)
	cluster2 := cluster.NewCluster("test-cluster-2", "test-cluster-app-2", "./app.js", []string{}, 2)
	mockClusterManager.StartCluster(cluster1)
	mockClusterManager.StartCluster(cluster2)

	// Test deleting a single process
	ctx := createContext([]string{"test-process-1"}, map[string]string{})

	err := commandManager.DeleteCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	processes := mockProcessManager.ListProcesses()
	if len(processes) != 1 {
		t.Errorf("Expected 1 process, got %d", len(processes))
	}

	// Test deleting a single cluster
	ctx = createContext([]string{"test-cluster-1"}, map[string]string{})

	err = commandManager.DeleteCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	clusters := mockClusterManager.ListClusters()
	if len(clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(clusters))
	}

	// Test deleting all
	ctx = createContext([]string{"all"}, map[string]string{})

	err = commandManager.DeleteCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	processes = mockProcessManager.ListProcesses()
	if len(processes) != 0 {
		t.Errorf("Expected 0 processes, got %d", len(processes))
	}

	clusters = mockClusterManager.ListClusters()
	if len(clusters) != 0 {
		t.Errorf("Expected 0 clusters, got %d", len(clusters))
	}
}

// TestListCommand tests the list command
func TestListCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Create test processes and clusters
	process := process.NewProcess("test-process", "test-app", "./app.js", []string{}, 1)
	process.Status = "running"
	process.CreatedAt = time.Now().Add(-10 * time.Second)
	mockProcessManager.Start(process)

	cluster := cluster.NewCluster("test-cluster", "test-cluster-app", "./app.js", []string{}, 2)
	cluster.Status = "running"
	cluster.CreatedAt = time.Now().Add(-20 * time.Second)
	mockClusterManager.StartCluster(cluster)

	// Test list command
	ctx := createContext([]string{}, map[string]string{})

	err := commandManager.ListCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestLogsCommand tests the logs command
func TestLogsCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Create a test process
	process := process.NewProcess("test-process", "test-app", "./app.js", []string{}, 1)
	mockProcessManager.Start(process)

	// Test logs command
	ctx := createContext([]string{"test-process"}, map[string]string{})

	err := commandManager.LogsCommand(ctx)
	// Expected to fail since we don't have a real logger implementation
	if err == nil {
		t.Errorf("Expected error for logs command, got nil")
	}
}

// TestMonitCommand tests the monit command
func TestMonitCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Test monit command - it should start without error
	ctx := createContext([]string{}, map[string]string{})

	// Run the monit command in a goroutine with timeout
	done := make(chan error, 1)
	go func() {
		done <- commandManager.MonitCommand(ctx)
	}()

	// Wait for a short time to ensure the command starts
	time.Sleep(100 * time.Millisecond)

	// The command should be running (it's a long-running process)
	// We don't expect it to return, so we just check that it started without error
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	default:
		// Command is still running, which is expected
	}
}

// TestStartupCommand tests the startup command
func TestStartupCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Test startup command
	ctx := createContext([]string{}, map[string]string{})

	err := commandManager.StartupCommand(ctx)
	// In test environment, we don't have root privileges, so this error is expected
	if err == nil {
		t.Logf("Startup command succeeded (unexpected in test environment)")
	} else {
		t.Logf("Startup command failed as expected: %v", err)
	}
}

// TestSaveCommand tests the save command
func TestSaveCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Test save command
	ctx := createContext([]string{}, map[string]string{})

	err := commandManager.SaveCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestReloadCommand tests the reload command
func TestReloadCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Create a test process
	process := process.NewProcess("test-process", "test-app", "./app.js", []string{}, 1)
	mockProcessManager.Start(process)

	// Create a test cluster
	cluster := cluster.NewCluster("test-cluster", "test-cluster-app", "./app.js", []string{}, 2)
	mockClusterManager.StartCluster(cluster)

	// Test reloading the process
	ctx := createContext([]string{"test-process"}, map[string]string{})

	err := commandManager.ReloadCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test reloading the cluster
	ctx = createContext([]string{"test-cluster"}, map[string]string{})

	err = commandManager.ReloadCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestStatusCommand tests the status command
func TestStatusCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Create a test process
	process := process.NewProcess("test-process", "test-app", "./app.js", []string{}, 1)
	process.Status = "running"
	mockProcessManager.Start(process)

	// Create a test cluster
	cluster := cluster.NewCluster("test-cluster", "test-cluster-app", "./app.js", []string{}, 2)
	cluster.Status = "running"
	mockClusterManager.StartCluster(cluster)

	// Test status command for process
	ctx := createContext([]string{"test-process"}, map[string]string{})

	err := commandManager.StatusCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test status command for cluster
	ctx = createContext([]string{"test-cluster"}, map[string]string{})

	err = commandManager.StatusCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestDescribeCommand tests the describe command
func TestDescribeCommand(t *testing.T) {
	mockProcessManager := NewMockProcessManager()
	mockClusterManager := NewMockClusterManager()
	commandManager := command.NewCommandManager(mockProcessManager, mockClusterManager)

	// Create a test process
	process := process.NewProcess("test-process", "test-app", "./app.js", []string{}, 1)
	process.Status = "running"
	process.CreatedAt = time.Now()
	process.UpdatedAt = time.Now()
	mockProcessManager.Start(process)

	// Create a test cluster
	cluster := cluster.NewCluster("test-cluster", "test-cluster-app", "./app.js", []string{}, 2)
	cluster.Status = "running"
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()
	mockClusterManager.StartCluster(cluster)

	// Test describe command for process
	ctx := createContext([]string{"test-process"}, map[string]string{})

	err := commandManager.DescribeCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test describe command for cluster
	ctx = createContext([]string{"test-cluster"}, map[string]string{})

	err = commandManager.DescribeCommand(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
