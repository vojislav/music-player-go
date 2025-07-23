package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Username   string `toml:"username"`
	Salt       string `toml:"salt"`
	Token      string `toml:"token"`
	Version    string `toml:"version"`
	ServerURL  string `toml:"server_url"`
	ClientName string `toml:"client_name"`
}

var config Config

// load config file into struct
func loadConfig(config *Config) {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
	}
	toml.Unmarshal(configData, config)
}

// write default and user-specfic config to file
func saveConfig() {
	defaultConfig, err := os.ReadFile("config")
	if err != nil {
		log.Fatalln(err)
	}

	// load default config fields to Config struct
	// username, token and salt are already added from loginForm
	toml.Unmarshal(defaultConfig, &config)

	fullConfig, err := toml.Marshal(config)
	if err != nil {
		printError(err)
	}

	_ = os.WriteFile(configFile, fullConfig, 0644)
}

func validConfig() bool {
	if _, err := os.Stat(configFile); err != nil {
		return false
	}

	var tempConfig Config
	loadConfig(&tempConfig)

	if tempConfig.Username == "" || tempConfig.Salt == "" ||
		tempConfig.Token == "" || tempConfig.ServerURL == "" ||
		tempConfig.Version == "" || tempConfig.ClientName == "" {
		return false
	}

	return true
}

func deleteConfig() {
	os.Remove(configFile)
}
