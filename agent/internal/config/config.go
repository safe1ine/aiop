package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerURL   string `yaml:"server_url"`
	EnrollToken string `yaml:"enroll_token"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		ServerURL:   getEnv("AIPO_SERVER_URL", "ws://localhost:8080/ws/agent"),
		EnrollToken: getEnv("AIPO_ENROLL_TOKEN", ""),
	}
	if path == "" {
		path = "agent.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if v := os.Getenv("AIPO_SERVER_URL"); v != "" {
		cfg.ServerURL = v
	}
	if v := os.Getenv("AIPO_ENROLL_TOKEN"); v != "" {
		cfg.EnrollToken = v
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
