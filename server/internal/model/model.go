package model

import "time"

type Agent struct {
	ID        int64     `json:"id"`
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	Status    string    `json:"status"` // online / offline
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        int64     `json:"id"`
	AgentID   int64     `json:"agent_id"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}
