package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type client struct {
	Name        string `yaml:"name"`
	ServiceName string `yaml:"serviceName"`
	BaseURL     string `yaml:"baseURL"`
	GQLFile     string `yaml:"gqlFile"`
	GoFile      string `yaml:"goFile"`
}

type config struct {
	PackageName string   `yaml:"packageName"`
	Clients     []client `yaml:"clients"`
}

func main() {
	configBytes, err := os.ReadFile("clients.yaml")
	if err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	var cfg config
	if err := yaml.Unmarshal(configBytes, &cfg); err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	if err := Generate(cfg); err != nil {
		log.Fatal(err)
	}
}
