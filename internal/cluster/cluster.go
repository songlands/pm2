package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"pm2/internal/log"
)

// Cluster represents a cluster of process instances
type Cluster struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Command      string    `json:"command"`
	Args         []string  `json:"args"`
	Instances    int       `json:"instances"`
	Status       string    `json:"status"` // running, stopped, errored
	PIDs         []int     `json:"pids"`
	RestartCount int       `json:"restart_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Logger       interface{} `json:"-"` // Logger interface, not serialized
	mutex        sync.RWMutex
}

// NewCluster creates a new cluster instance
func NewCluster(id, name, command string, args []string, instances int) *Cluster {
	return &Cluster{
		ID:           id,
		Name:         name,
		Command:      command,
		Args:         args,
		Instances:    instances,
		Status:       "stopped",
		PIDs:         []int{},
		RestartCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Start starts all instances in the cluster
func (c *Cluster) Start() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Status == "running" {
		return fmt.Errorf("cluster is already running")
	}

	// Create logger
	logger, err := log.NewLogger(c.ID)
	if err != nil {
		return err
	}
	c.Logger = logger

	c.PIDs = []int{}

	for i := 0; i < c.Instances; i++ {
		cmd := exec.Command(c.Command, c.Args...)
		cmd.Stdout = logger.Stdout()
		cmd.Stderr = logger.Stderr()

		if err := cmd.Start(); err != nil {
			logger.Close()
			return err
		}

		c.PIDs = append(c.PIDs, cmd.Process.Pid)
	}

	c.Status = "running"
	c.UpdatedAt = time.Now()

	return nil
}

// Stop stops all instances in the cluster
func (c *Cluster) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Status != "running" {
		return fmt.Errorf("cluster is not running")
	}

	for _, pid := range c.PIDs {
		process, err := os.FindProcess(pid)
		if err == nil {
			if err := process.Kill(); err != nil {
				return err
			}
		}
	}

	c.Status = "stopped"
	c.PIDs = []int{}
	c.UpdatedAt = time.Now()

	return nil
}

// Restart restarts all instances in the cluster
func (c *Cluster) Restart() error {
	if err := c.Stop(); err != nil {
		return err
	}

	c.mutex.Lock()
	// Increment restart count
	c.RestartCount++
	c.mutex.Unlock()

	return c.Start()
}

// Reload reloads the cluster with zero downtime
func (c *Cluster) Reload() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Status != "running" {
		return fmt.Errorf("cluster is not running")
	}

	newPIDs := []int{}

	// Start new instances one by one
	for i := 0; i < c.Instances; i++ {
		cmd := exec.Command(c.Command, c.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			// Clean up any started instances
			for _, pid := range newPIDs {
				process, err := os.FindProcess(pid)
				if err == nil {
					process.Kill()
				}
			}
			return err
		}

		newPIDs = append(newPIDs, cmd.Process.Pid)

		// Wait a bit for the new instance to start
		time.Sleep(500 * time.Millisecond)

		// Stop the corresponding old instance
		if i < len(c.PIDs) {
			process, err := os.FindProcess(c.PIDs[i])
			if err == nil {
				process.Kill()
			}
		}
	}

	// Stop any remaining old instances
	for i := c.Instances; i < len(c.PIDs); i++ {
		process, err := os.FindProcess(c.PIDs[i])
		if err == nil {
			process.Kill()
		}
	}

	c.PIDs = newPIDs
	c.UpdatedAt = time.Now()

	return nil
}
