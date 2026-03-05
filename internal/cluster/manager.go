package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Manager handles cluster lifecycle management
type Manager struct {
	clusters map[string]*Cluster
	mutex    sync.RWMutex
	configPath string
}

// NewManager creates a new cluster manager
func NewManager() *Manager {
	configPath := filepath.Join(os.TempDir(), "pm2", "clusters.json")
	manager := &Manager{
		clusters: make(map[string]*Cluster),
		configPath: configPath,
	}

	// Load clusters from config
	manager.Load()

	return manager
}

// Load loads clusters from config file
func (m *Manager) Load() error {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var clusters map[string]*Cluster
	if err := json.Unmarshal(data, &clusters); err != nil {
		return err
	}

	m.clusters = clusters
	return nil
}

// Save saves clusters to config file
func (m *Manager) Save() error {
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(m.clusters)
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// StartCluster starts a cluster
func (m *Manager) StartCluster(c *Cluster) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.clusters[c.ID]; exists {
		return fmt.Errorf("cluster with ID %s already exists", c.ID)
	}

	if err := c.Start(); err != nil {
		return err
	}

	m.clusters[c.ID] = c

	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save clusters: %v\n", err)
	}

	return nil
}

// StopCluster stops a cluster
func (m *Manager) StopCluster(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, exists := m.clusters[id]
	if !exists {
		return fmt.Errorf("cluster with ID %s not found", id)
	}

	if err := c.Stop(); err != nil {
		return err
	}

	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save clusters: %v\n", err)
	}

	return nil
}

// RestartCluster restarts a cluster
func (m *Manager) RestartCluster(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, exists := m.clusters[id]
	if !exists {
		return fmt.Errorf("cluster with ID %s not found", id)
	}

	if err := c.Restart(); err != nil {
		return err
	}

	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save clusters: %v\n", err)
	}

	return nil
}

// DeleteCluster removes a cluster
func (m *Manager) DeleteCluster(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, exists := m.clusters[id]
	if !exists {
		return fmt.Errorf("cluster with ID %s not found", id)
	}

	// Stop the cluster if it's running
	if c.Status == "running" {
		// Try to stop, but continue even if it fails
		if err := c.Stop(); err != nil {
			fmt.Printf("WARNING: Failed to stop cluster %s: %v\n", id, err)
		}
	}

	// Close logger
	if c.Logger != nil {
		if logger, ok := c.Logger.(interface{ Close() error }); ok {
			logger.Close()
		}
	}

	delete(m.clusters, id)

	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save clusters: %v\n", err)
	}

	return nil
}

// ReloadCluster reloads a cluster with zero downtime
func (m *Manager) ReloadCluster(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, exists := m.clusters[id]
	if !exists {
		return fmt.Errorf("cluster with ID %s not found", id)
	}

	if err := c.Reload(); err != nil {
		return err
	}

	if err := m.Save(); err != nil {
		fmt.Printf("WARNING: Failed to save clusters: %v\n", err)
	}

	return nil
}

// GetCluster returns a cluster by ID
func (m *Manager) GetCluster(id string) (*Cluster, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, exists := m.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster with ID %s not found", id)
	}

	return c, nil
}

// ListClusters returns all clusters
func (m *Manager) ListClusters() []*Cluster {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	clusters := make([]*Cluster, 0, len(m.clusters))
	for _, c := range m.clusters {
		clusters = append(clusters, c)
	}

	return clusters
}
