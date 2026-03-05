package process

import (
	"time"
)

// Process represents a managed application process
type Process struct {
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
}

// NewProcess creates a new process instance
func NewProcess(id, name, command string, args []string, instances int) *Process {
	return &Process{
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
