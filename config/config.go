package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Default     string
		Connections map[string]map[string]string
	}
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
}

var AppConfig *Config

func LoadConfig(path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Config file read error: %v", err)
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("YAML parse error: %v", err)
	}

	AppConfig = &cfg
}
