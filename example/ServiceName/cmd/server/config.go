package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
)

// sets up the config parser
func setup() {
	viper.SetConfigName(serviceName)
	viper.AddConfigPath("/opt/config/")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	setConfigDefaults()
}

func readConfig() {
	setup()
	logger.Log("msg", "Reading config")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		// Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	data, _ := json.MarshalIndent(viper.AllSettings(), "", "  ")
	logger.Log("Config", string(data))
}
