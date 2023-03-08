package orion

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"

	"github.com/carousell/Orion/v2/utils/log"
)

var (
	configPaths = []string{".", "/opt/config/"}
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
	//Env is the environment this service is running in
	Env string
	// DefaultJSONPB sets jsonpb as the encoder/decoder for application/json request/response bodies
	DefaultJSONPB bool
	// DisableDefaultInterceptors disables the default interceptors for all handlers
	DisableDefaultInterceptors bool
	// Receive message Size is used to update the default limit of message that can be received
	MaxRecvMsgSize int
}

//BuildDefaultConfig builds a default config object for Orion
func BuildDefaultConfig(name string) Config {
	setup(name)
	readConfig(name)
	return Config{
		GRPCOnly:                   viper.GetBool("orion.GRPCOnly"),
		HTTPOnly:                   viper.GetBool("orion.HTTPOnly"),
		GRPCPort:                   viper.GetString("orion.GRPCPort"),
		HTTPPort:                   viper.GetString("orion.HTTPPort"),
		Env:                        viper.GetString("orion.Env"),
		OrionServerName:            name,
		DefaultJSONPB:              viper.GetBool("orion.DefaultJSONPB"),
		DisableDefaultInterceptors: viper.GetBool("orion.DisableDefaultInterceptors"),
		MaxRecvMsgSize:             viper.GetInt("orion.MaxRecvMsgSize"),
	}
}

func setConfigDefaults() {
	viper.SetDefault("orion.GRPCPort", "9281")
	viper.SetDefault("orion.HttpPort", "9282")
	viper.SetDefault("orion.GRPCOnly", false)
	viper.SetDefault("orion.HTTPOnly", false)
	viper.SetDefault("orion.env", "dev")
	viper.SetDefault("orion.Env", "development")
	viper.SetDefault("orion.DefaultJSONPB", false)
	viper.SetDefault("orion.MaxRecvMsgSize", -1)

}

// sets up the config parser
func setup(name string) {
	viper.SetConfigName(name)
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}
	viper.AutomaticEnv()
	setConfigDefaults()
}

func readConfig(name string) error {
	ctx := context.Background()
	log.Info(ctx, "config", "Reading config")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		// do nothing and default everything
		log.Warn(ctx, "config", "config could not be read "+err.Error())
		return fmt.Errorf("Config config could not be read %s", err.Error())
	}
	data, _ := json.MarshalIndent(viper.AllSettings(), "", "  ")
	log.Info(ctx, "Config", string(data))
	return nil
}

// AddConfigPath adds a config path from where orion tries to read config values
func AddConfigPath(path ...string) {
	if configPaths == nil {
		configPaths = []string{}
	}
	configPaths = append(configPaths, path...)
}

// ResetConfigPath resets the configuration paths
func ResetConfigPath() {
	configPaths = []string{}
}
