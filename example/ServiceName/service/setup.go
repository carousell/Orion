package service

import "github.com/spf13/viper"

func SetSvcDefaults() {
	viper.SetDefault("main-app.timeout", 500)
}

func BuildSvcConfig() Config {
	return Config{
		// cassandra
		CasKeyspace:         viper.GetString("storage-cas.keyspace"),
		CasHosts:            viper.GetStringSlice("storage-cas.hosts"),
		CasConsistency:      viper.GetString("storage-cas.consistency"),
		CasConnectTimeout:   viper.GetInt("storage-cas.connectTimeout"),
		CasOperationTimeout: viper.GetInt("storage-cas.operationTimeout"),
		CasConnections:      viper.GetInt("storage-cas.connections"),
		// es
		ESUrl:         viper.GetString("storage-es.url"),
		ESPrefix:      viper.GetString("storage-es.prefix"),
		ESFakeContext: viper.GetBool("storage-es.fakeContext"),
	}
}
