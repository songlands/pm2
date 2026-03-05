package process

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"pm2/internal/log"
)

// Manager handles process lifecycle management
type Manager struct {
	processes map[string]*Process
	mutex     sync.RWMutex
	configPath string
}

// NewManager creates a new process manager
func NewManager() *Manager {
	configPath := filepath.Join(os.TempDir(), "pm2", "processes.json")
	manager := &Manager{
		processes: make(map[string]*Process),
		configPath: configPath,
	}

	// Load processes from config
	manager.Load()

	return manager
}

// Load loads processes from config file
func (m *Manager) Load() error {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var processes map[string]*Process
	if err := json.Unmarshal(data, &processes); err != nil {
		return err
	}

	m.processes = processes
	return nil
}

// Save saves processes to config file
func (m *Manager) Save() error {
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(m.processes)
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Start starts a process
func (m *Manager) Start(p *Process) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.processes[p.ID]; exists {
		return fmt.Errorf("process with ID %s already exists", p.ID)
	}

	// Create logger
	logger, err := log.NewLogger(p.ID)
	if err != nil {
		return err
	}
	p.Logger = logger

	// Start the process
	cmd := exec.Command(p.Command, p.Args...)
	cmd.Stdout = logger.Stdout()
	cmd.Stderr = logger.Stderr()

	if err := cmd.Start(); err != nil {
		logger.Close()
		return err
	}

	p.PIDs = append(p.PIDs, cmd.Process.Pid)
	p.Status = "running"
	p.UpdatedAt = time.Now()

	m.processes[p.ID] = p
	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save processes: %v\n", err)
	}

	return nil
}

// Stop stops a process
func (m *Manager) Stop(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	p, exists := m.processes[id]
	if !exists {
		return fmt.Errorf("process with ID %s not found", id)
	}

	for _, pid := range p.PIDs {
		process, err := os.FindProcess(pid)
		if err == nil {
			if err := process.Kill(); err != nil {
				return err
			}
		}
	}

	p.Status = "stopped"
	p.PIDs = []int{}
	p.UpdatedAt = time.Now()

	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save processes: %v\n", err)
	}

	return nil
}

// Restart restarts a process
func (m *Manager) Restart(id string) error {
	m.mutex.Lock()
	p, exists := m.processes[id]
	if !exists {
		m.mutex.Unlock()
		return fmt.Errorf("process with ID %s not found", id)
	}

	// Increment restart count
	p.RestartCount++
	m.mutex.Unlock()

	if err := m.Stop(id); err != nil {
		return err
	}

	m.mutex.Lock()
	p, exists = m.processes[id]
	if !exists {
		m.mutex.Unlock()
		return fmt.Errorf("process with ID %s not found", id)
	}
	m.mutex.Unlock()

	// Create a new process with the same parameters
	newP := NewProcess(p.ID, p.Name, p.Command, p.Args, p.Instances)
	newP.RestartCount = p.RestartCount

	m.mutex.Lock()
	delete(m.processes, id)
	m.mutex.Unlock()

	return m.Start(newP)
}

// Delete removes a process
func (m *Manager) Delete(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	p, exists := m.processes[id]
	if !exists {
		return fmt.Errorf("process with ID %s not found", id)
	}

	// Stop the process if it's running
	if p.Status == "running" {
		for _, pid := range p.PIDs {
			process, err := os.FindProcess(pid)
			if err == nil {
				process.Kill()
			}
		}
	}

	// Close logger
	if p.Logger != nil {
		if logger, ok := p.Logger.(interface{ Close() error }); ok {
			logger.Close()
		}
	}

	delete(m.processes, id)

	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save processes: %v\n", err)
	}

	return nil
}

// Reload reloads a process with zero downtime
func (m *Manager) Reload(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	p, exists := m.processes[id]
	if !exists {
		return fmt.Errorf("process with ID %s not found", id)
	}

	// Start new process
	newP := NewProcess(fmt.Sprintf("%s_reload_%d", p.ID, time.Now().Unix()), p.Name, p.Command, p.Args, p.Instances)
	if err := m.Start(newP); err != nil {
		return err
	}

	// Wait a bit for the new process to start
	time.Sleep(1 * time.Second)

	// Stop the old process
	if err := m.Stop(id); err != nil {
		// If stopping fails, try to stop the new process
		m.Stop(newP.ID)
		return err
	}

	// Replace the old process with the new one
	delete(m.processes, id)
	newP.ID = id
	m.processes[id] = newP

	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save processes: %v\n", err)
	}

	return nil
}

// GetProcess returns a process by ID
func (m *Manager) GetProcess(id string) (*Process, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	p, exists := m.processes[id]
	if !exists {
		return nil, fmt.Errorf("process with ID %s not found", id)
	}

	return p, nil
}

// ListProcesses returns all processes
func (m *Manager) ListProcesses() []*Process {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	processes := make([]*Process, 0, len(m.processes))
	for _, p := range m.processes {
		processes = append(processes, p)
	}

	return processes
}
