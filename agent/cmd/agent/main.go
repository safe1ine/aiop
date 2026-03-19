package main

import (
	"flag"
	"log"

	"github.com/aipo/agent/internal/config"
	"github.com/aipo/agent/internal/connector"
)

func main() {
	cfgPath := flag.String("config", "agent.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.EnrollToken == "" {
		log.Fatal("enroll token is required (AIPO_ENROLL_TOKEN or agent.yaml)")
	}

	log.Printf("starting agent -> %s", cfg.ServerURL)
	client := connector.New(cfg)
	client.Run()
}
