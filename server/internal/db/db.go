package db

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/aipo/server/internal/model"
	_ "modernc.org/sqlite"
)

//go:embed migrate/*.sql
var migrations embed.FS

type Store interface {
	// User
	GetUserByUsername(username string) (*model.User, error)
	CreateUser(username, passwordHash string) error

	// Agent — self-registers on connect, identified by hostname+ip
	ListAgents() ([]*model.Agent, error)
	GetAgent(id int64) (*model.Agent, error)
	UpsertAgent(hostname, ip, os, arch string) (*model.Agent, error) // create or update
	UpdateAgentStatus(id int64, status string) error
	UpdateAgentLastSeen(id int64) error
	DeleteAgent(id int64) error

	// Audit
	CreateAuditLog(agentID int64, action, detail string) error
}

type sqliteStore struct {
	db *sql.DB
}

func Open(path string) (Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := runMigrations(db); err != nil {
		return nil, err
	}
	return &sqliteStore{db: db}, nil
}

func runMigrations(db *sql.DB) error {
	entries, err := migrations.ReadDir("migrate")
	if err != nil {
		return err
	}
	for _, e := range entries {
		data, err := migrations.ReadFile("migrate/" + e.Name())
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("migration %s: %w", e.Name(), err)
		}
	}
	return nil
}

func (s *sqliteStore) GetUserByUsername(username string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (s *sqliteStore) CreateUser(username, passwordHash string) error {
	_, err := s.db.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`, username, passwordHash,
	)
	return err
}

func (s *sqliteStore) ListAgents() ([]*model.Agent, error) {
	rows, err := s.db.Query(
		`SELECT id, hostname, ip, os, arch, status, last_seen, created_at FROM agents ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	agents := make([]*model.Agent, 0)
	for rows.Next() {
		a := &model.Agent{}
		var lastSeen sql.NullTime
		if err := rows.Scan(&a.ID, &a.Hostname, &a.IP, &a.OS, &a.Arch, &a.Status, &lastSeen, &a.CreatedAt); err != nil {
			return nil, err
		}
		if lastSeen.Valid {
			a.LastSeen = lastSeen.Time
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

func (s *sqliteStore) GetAgent(id int64) (*model.Agent, error) {
	a := &model.Agent{}
	var lastSeen sql.NullTime
	err := s.db.QueryRow(
		`SELECT id, hostname, ip, os, arch, status, last_seen, created_at FROM agents WHERE id = ?`, id,
	).Scan(&a.ID, &a.Hostname, &a.IP, &a.OS, &a.Arch, &a.Status, &lastSeen, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if lastSeen.Valid {
		a.LastSeen = lastSeen.Time
	}
	return a, err
}

// UpsertAgent finds existing agent by hostname+ip or creates a new one, then updates info.
func (s *sqliteStore) UpsertAgent(hostname, ip, os, arch string) (*model.Agent, error) {
	a := &model.Agent{}
	var lastSeen sql.NullTime
	err := s.db.QueryRow(
		`SELECT id, hostname, ip, os, arch, status, last_seen, created_at FROM agents WHERE hostname = ? AND ip = ?`,
		hostname, ip,
	).Scan(&a.ID, &a.Hostname, &a.IP, &a.OS, &a.Arch, &a.Status, &lastSeen, &a.CreatedAt)

	if err == sql.ErrNoRows {
		// New agent
		res, err := s.db.Exec(
			`INSERT INTO agents (hostname, ip, os, arch, status, last_seen) VALUES (?, ?, ?, ?, 'online', ?)`,
			hostname, ip, os, arch, time.Now(),
		)
		if err != nil {
			return nil, err
		}
		id, _ := res.LastInsertId()
		return &model.Agent{ID: id, Hostname: hostname, IP: ip, OS: os, Arch: arch, Status: "online", LastSeen: time.Now(), CreatedAt: time.Now()}, nil
	}
	if err != nil {
		return nil, err
	}

	// Existing agent — update info
	_, err = s.db.Exec(
		`UPDATE agents SET os=?, arch=?, status='online', last_seen=? WHERE id=?`,
		os, arch, time.Now(), a.ID,
	)
	if err != nil {
		return nil, err
	}
	a.OS = os
	a.Arch = arch
	a.Status = "online"
	a.LastSeen = time.Now()
	return a, nil
}

func (s *sqliteStore) UpdateAgentStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE agents SET status=? WHERE id=?`, status, id)
	return err
}

func (s *sqliteStore) UpdateAgentLastSeen(id int64) error {
	_, err := s.db.Exec(`UPDATE agents SET last_seen=? WHERE id=?`, time.Now(), id)
	return err
}

func (s *sqliteStore) DeleteAgent(id int64) error {
	_, err := s.db.Exec(`DELETE FROM agents WHERE id=?`, id)
	return err
}

func (s *sqliteStore) CreateAuditLog(agentID int64, action, detail string) error {
	_, err := s.db.Exec(
		`INSERT INTO audit_logs (agent_id, action, detail) VALUES (?, ?, ?)`, agentID, action, detail,
	)
	return err
}
