package orion

func BuildDefaultConfig() Config {
	return Config{
		GRPCOnly:             false,
		GRPCPort:             9891,
		HTTPPort:             9892,
		ReloadOnConfigChange: true,
		OrionServerName:      "MainOrionService",
	}
}
