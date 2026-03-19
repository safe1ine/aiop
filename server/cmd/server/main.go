package main

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"github.com/aipo/server/internal/api"
	"github.com/aipo/server/internal/config"
	"github.com/aipo/server/internal/db"
	"github.com/aipo/server/internal/hub"
)

func main() {
	cfg := config.Load()

	store, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	// Ensure admin user exists
	if err := ensureAdmin(store, cfg.AdminUser, cfg.AdminPass); err != nil {
		log.Fatalf("admin: %v", err)
	}

	h := hub.New()
	router := api.NewRouter(store, h, cfg.JWTSecret, cfg.AdminUser, cfg.AdminPass, cfg.EnrollToken)

	log.Printf("server listening on %s", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func ensureAdmin(store db.Store, username, password string) error {
	user, err := store.GetUserByUsername(username)
	if err != nil {
		return err
	}
	if user != nil {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return store.CreateUser(username, string(hash))
}
