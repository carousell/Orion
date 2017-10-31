package orion

import (
	"encoding/json"
	"log"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/spf13/viper"
)

// Config is the configuration used by Orion core
type Config struct {
	//OrionServerName is the name of this orion server that is tracked
	OrionServerName string
	// GRPCOnly tells orion not to build HTTP/1.1 server and only initializes gRPC server
	GRPCOnly bool
	//HTTPOnly tells orion not to build gRPC server and only initializes HTTP/1.1 server
	HTTPOnly bool
	// HTTPPort is the port to bind for HTTP requests
	HTTPPort string
	// GRPCPost id the port to bind for gRPC requests
	GRPCPort string
	// ReloadOnConfigChange when set reloads the service when it detects configuration update
	ReloadOnConfigChange bool
	//HystrixConfig is the configuration options for hystrix
	HystrixConfig HystrixConfig
	//ZipkinConfig is the configuration options for zipkin
	ZipkinConfig ZipkinConfig
	//NewRelicConfig is the configuration options for new relic
	NewRelicConfig NewRelicConfig
}

// HystrixConfig is configuration used by hystrix
type HystrixConfig struct {
	//Port is the port to start hystrix stream handler on
	Port string
	//CommandConfig is configuration for individual commands
	CommandConfig map[string]hystrix.CommandConfig
}

//ZipkinConfig is the configuration for the zipkin collector
type ZipkinConfig struct {
	//Addr is the address of the zipkin collector
	Addr string
}

//NewRelicConfig is the configuration for newrelic
type NewRelicConfig struct {
	APIKey      string
	ServiceName string
}

//BuildDefaultConfig builds a default config object for Orion
func BuildDefaultConfig(name string) Config {
	readConfig(name)
	return Config{
		GRPCOnly:             false,
		HTTPOnly:             false,
		GRPCPort:             viper.GetString("orion.GRPCPort"),
		HTTPPort:             viper.GetString("orion.HTTPPort"),
		ReloadOnConfigChange: true,
		OrionServerName:      name,
		HystrixConfig:        BuildDefaultHystrixConfig(),
		ZipkinConfig:         BuildDefaultZipkinConfig(),
		NewRelicConfig:       BuildDefaultNewRelicConfig(),
	}
}

//BuildDefaultHystrixConfig builds a default config for hystrix
func BuildDefaultHystrixConfig() HystrixConfig {
	return HystrixConfig{
		Port:          viper.GetString("orion.HystrixPort"),
		CommandConfig: make(map[string]hystrix.CommandConfig),
	}
}

//BuildDefaultZipkinConfig builds a default config for zipkin
func BuildDefaultZipkinConfig() ZipkinConfig {
	return ZipkinConfig{
		Addr: viper.GetString("orion.ZipkinAddr"),
	}
}

//BuildDefaultNewRelicConfig builds a default config for newrelic
func BuildDefaultNewRelicConfig() NewRelicConfig {
	return NewRelicConfig{
		APIKey:      viper.GetString("orion.newrelic-servicename"),
		ServiceName: viper.GetString("orion.newrelic-api-key"),
	}
}

func setConfigDefaults() {
	viper.SetDefault("orion.ServiceName", "MainOrionService")
	viper.SetDefault("orion.GRPCPort", "9281")
	viper.SetDefault("orion.HttpPort", "9282")
	viper.SetDefault("orion.HystrixPort", "9283")
	viper.SetDefault("orion.PprofPort", "9284")
	viper.SetDefault("orion.ZipkinAddr", "http://10.200.0.7:9411/api/v1/spans")
	viper.SetDefault("orion.env", "dev")
	viper.SetDefault("orion.rollbar-token", "")
	viper.SetDefault("orion.newrelic-servicename", "")
	viper.SetDefault("orion.newrelic-api-key", "")
}

// sets up the config parser
func setup(name string) {
	viper.SetConfigName(name)
	viper.AddConfigPath("/opt/config/")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	setConfigDefaults()
}

func readConfig(name string) {
	setup(name)
	log.Println("msg", "Reading config")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		// do nothing and default everything
		log.Println("Config", "config could not be read "+err.Error())
		return
	}
	data, _ := json.MarshalIndent(viper.AllSettings(), "", "  ")
	log.Println("Config", string(data))
}
