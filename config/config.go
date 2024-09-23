package config

import (
	"encoding/json"
	"log"
	"os"
)

type AppConfig struct {
	Id    string `json:"id"`
	Token string `json:"token"`
	Realm string `json:"realm"`
}

type SourceTargetConfig struct {
	Source AppConfig `json:"source"`
	Target AppConfig `json:"target"`
}

type Config struct {
	Source AppConfig `json:"source"`
	Target AppConfig `json:"target"`
	Pages  []int     `json:"pages"`
}

var defaultConfig Config = Config{
	Source: AppConfig{
		Id:    "",
		Token: "",
		Realm: "",
	},
	Target: AppConfig{
		Id:    "",
		Token: "",
		Realm: "",
	},
	Pages: []int{},
}

func createConfig() Config {
	configFile, err := os.Create("config.json")

	if err != nil {
		log.Fatal(err)
	}

	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(defaultConfig)

	if err != nil {
		log.Fatal(err)
	}

	return defaultConfig
}

func ReadConfig() Config {
	if _, err := os.Stat("config.json"); err == nil {
		configFile, err := os.ReadFile("config.json")

		if err != nil {
			log.Fatal(err)
		}

		var config Config

		json.Unmarshal(configFile, &config)

		return config
	} else if os.IsNotExist(err) {
		createConfig()
	} else {
		log.Fatal(err)
	}

	return defaultConfig
}
