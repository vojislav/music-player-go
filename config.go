package main

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

var client_name = "music-player-go"
var version = "1.16.1"
var server_url = "https://music.lazic.xyz/rest/"

type Config struct {
	Username  string `toml:"username"`
	Salt      string `toml:"salt"`
	Token     string `toml:"token"`
	Version   string `toml:"version"`
	ServerURL string `toml:"server_url"`
}

var config Config

func readConfig() {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
	}
	toml.Unmarshal(configData, &config)
}

func writeConfig() {
	config.ServerURL = server_url
	config.Version = version
	configToml, _ := toml.Marshal(config)
	_ = os.WriteFile(configFile, configToml, 0644)
}

func validConfig() bool {
	if _, err := os.Stat(configFile); err != nil {
		return false
	}

	readConfig()

	if config.Username == "" || config.Salt == "" ||
		config.Token == "" || config.ServerURL == "" ||
		config.Version == "" {
		return false
	}

	return true
}

func deleteConfig() {
	os.Remove(configFile)
}
