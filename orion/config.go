package orion

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/spf13/viper"

	"github.com/carousell/Orion/utils/log"
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
	//PprofPort is the port to use for pprof
	PProfport string
	// HotReload when set reloads the service when it receives SIGHUP
	HotReload bool
	//EnableProtoURL adds gRPC generated urls in HTTP handler
	EnableProtoURL bool
	//EnablePrometheus enables prometheus metric for services on path '/metrics' on pprof port
	EnablePrometheus bool
	//EnablePrometheusHistograms enables request histograms for services
	//ref: https://github.com/grpc-ecosystem/go-grpc-prometheus#histograms
	EnablePrometheusHistogram bool
	//HystrixConfig is the configuration options for hystrix
	HystrixConfig HystrixConfig
	//ZipkinConfig is the configuration options for zipkin
	ZipkinConfig ZipkinConfig
	//NewRelicConfig is the configuration options for new relic
	NewRelicConfig NewRelicConfig
	//RollbarToken is the token to be used in rollbar
	RollbarToken string
	//SentryDSN is the token used by sentry for error reporting
	SentryDSN string
	//Env is the environment this service is running in
	Env string
	// DefaultJSONPB sets jsonpb as the encoder/decoder for application/json request/response bodies
	DefaultJSONPB bool
	// DisableDefaultInterceptors disables the default interceptors for all handlers
	DisableDefaultInterceptors bool
	// Receive message Size is used to update the default limit of message that can be received
	MaxRecvMsgSize int
	// Read timeout http server, The time in seconds read request body timeout.
	ReadTimeout int
	// Write timeout http server, The time in seconds that start from read request complete to write the resp to client must be happened in this time.
	WriteTimeout int
}

// HystrixConfig is configuration used by hystrix
type HystrixConfig struct {
	//Port is the port to start hystrix stream handler on
	Port string
	//CommandConfig is configuration for individual commands
	CommandConfig map[string]hystrix.CommandConfig
	//StatsdAddr is the address of the statsd hosts to send hystrix data to
	StatsdAddr string
	// DefaultTimeout is how long to wait for command to complete, in milliseconds
	DefaultTimeout int
	// DefaultMaxConcurrent is how many commands of the same type can run at the same time
	DefaultMaxConcurrent int
	// DefaultVolumeThreshold is the minimum number of requests needed before a circuit can be tripped due to health
	DefaultVolumeThreshold int
	// DefaultSleepWindow is how long, in milliseconds, to wait after a circuit opens before testing for recovery
	DefaultSleepWindow int
	// DefaultErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
	DefaultErrorPercentThreshold int
}

// ZipkinConfig is the configuration for the zipkin collector
type ZipkinConfig struct {
	//Addr is the address of the zipkin collector
	Addr string
}

// NewRelicConfig is the configuration for newrelic
type NewRelicConfig struct {
	APIKey      string
	ServiceName string
	//HttpTxNameType decides the transaction name logged in NR. Options are "fullmethod" (default), "method" , "route".
	HttpTxNameType    string
	IncludeAttributes []string
	ExcludeAttributes []string
}

// BuildDefaultConfig builds a default config object for Orion
func BuildDefaultConfig(name string) Config {
	setup(name)
	readConfig(name)
	return Config{
		GRPCOnly:                   viper.GetBool("orion.GRPCOnly"),
		HTTPOnly:                   viper.GetBool("orion.HTTPOnly"),
		GRPCPort:                   viper.GetString("orion.GRPCPort"),
		HTTPPort:                   viper.GetString("orion.HTTPPort"),
		PProfport:                  viper.GetString("orion.PprofPort"),
		HotReload:                  viper.GetBool("orion.HotReload"),
		EnableProtoURL:             viper.GetBool("orion.EnableProtoURL"),
		EnablePrometheus:           viper.GetBool("orion.EnablePrometheus"),
		EnablePrometheusHistogram:  viper.GetBool("orion.EnablePrometheusHistogram"),
		RollbarToken:               viper.GetString("orion.rollbar-token"),
		Env:                        viper.GetString("orion.Env"),
		SentryDSN:                  viper.GetString("orion.SentryDSN"),
		OrionServerName:            name,
		HystrixConfig:              BuildDefaultHystrixConfig(),
		ZipkinConfig:               BuildDefaultZipkinConfig(),
		NewRelicConfig:             BuildDefaultNewRelicConfig(),
		DefaultJSONPB:              viper.GetBool("orion.DefaultJSONPB"),
		DisableDefaultInterceptors: viper.GetBool("orion.DisableDefaultInterceptors"),
		MaxRecvMsgSize:             viper.GetInt("orion.MaxRecvMsgSize"),
		ReadTimeout:                viper.GetInt("orion.ReadTimeout"),
		WriteTimeout:               viper.GetInt("orion.WriteTimeout"),
	}
}

// BuildDefaultHystrixConfig builds a default config for hystrix
func BuildDefaultHystrixConfig() HystrixConfig {
	return HystrixConfig{
		Port:                         viper.GetString("orion.HystrixPort"),
		CommandConfig:                make(map[string]hystrix.CommandConfig),
		StatsdAddr:                   viper.GetString("orion.HystrixStatsd"),
		DefaultTimeout:               viper.GetInt("orion.HystrixDefaultTimeout"),
		DefaultMaxConcurrent:         viper.GetInt("orion.HystrixDefaultMaxConcurrent"),
		DefaultVolumeThreshold:       viper.GetInt("orion.HystrixDefaultVolumeThreshold"),
		DefaultSleepWindow:           viper.GetInt("orion.HystrixDefaultSleepWindow"),
		DefaultErrorPercentThreshold: viper.GetInt("orion.HystrixDefaultErrorPercentThreshold"),
	}
}

// BuildDefaultZipkinConfig builds a default config for zipkin
func BuildDefaultZipkinConfig() ZipkinConfig {
	return ZipkinConfig{
		Addr: viper.GetString("orion.ZipkinAddr"),
	}
}

// BuildDefaultNewRelicConfig builds a default config for newrelic
func BuildDefaultNewRelicConfig() NewRelicConfig {
	return NewRelicConfig{
		ServiceName:       viper.GetString("orion.NewRelicServiceName"),
		APIKey:            viper.GetString("orion.NewRelicApiKey"),
		HttpTxNameType:    viper.GetString("orion.NewRelicHttpTxNameType"),
		ExcludeAttributes: viper.GetStringSlice("orion.NewRelicExclude"),
		IncludeAttributes: viper.GetStringSlice("orion.NewRelicInclude"),
	}
}

func setConfigDefaults() {
	viper.SetDefault("orion.GRPCPort", "9281")
	viper.SetDefault("orion.HttpPort", "9282")
	viper.SetDefault("orion.HystrixPort", "9283")
	viper.SetDefault("orion.PprofPort", "9284")
	viper.SetDefault("orion.GRPCOnly", false)
	viper.SetDefault("orion.HTTPOnly", false)
	viper.SetDefault("orion.EnableProtoURL", false)
	viper.SetDefault("orion.ZipkinAddr", "")
	viper.SetDefault("orion.env", "dev")
	viper.SetDefault("orion.rollbar-token", "")
	viper.SetDefault("orion.HotReload", true)
	viper.SetDefault("orion.EnablePrometheus", true)
	viper.SetDefault("orion.EnablePrometheusHistogram", false)
	viper.SetDefault("orion.Env", "development")
	viper.SetDefault("orion.DefaultJSONPB", false)

	viper.SetDefault("orion.HystrixDefaultTimeout", 1000)
	viper.SetDefault("orion.HystrixDefaultMaxConcurrent", 300)
	viper.SetDefault("orion.HystrixDefaultVolumeThreshold", 75)
	viper.SetDefault("orion.HystrixDefaultSleepWindow", 1000)
	viper.SetDefault("orion.HystrixDefaultErrorPercentThreshold", 75)
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
