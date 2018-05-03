package worker

import (
	"fmt"
	"strings"
)

func getServerName(config Config) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", config.RabbitConfig.UserName, config.RabbitConfig.Password,
		config.RabbitConfig.Host, config.RabbitConfig.Port, config.RabbitConfig.BrokerVHost)
}

func buildRabbitConfig(cfg Config) *RabbitMQConfig {
	if cfg.RabbitConfig != nil {
		rabbitConfig := new(RabbitMQConfig)
		if strings.TrimSpace(cfg.RabbitConfig.BrokerVHost) != "" {
			rabbitConfig.BrokerVHost = cfg.RabbitConfig.BrokerVHost
		} else {
			rabbitConfig.BrokerVHost = "workers"
		}

		if strings.TrimSpace(cfg.RabbitConfig.Host) != "" {
			rabbitConfig.Host = cfg.RabbitConfig.Host
		} else {
			rabbitConfig.Host = "localhost"
		}

		if strings.TrimSpace(cfg.RabbitConfig.Password) != "" {
			rabbitConfig.Password = cfg.RabbitConfig.Password
		} else {
			rabbitConfig.Host = "guest"
		}

		if strings.TrimSpace(cfg.RabbitConfig.Port) != "" {
			rabbitConfig.Port = cfg.RabbitConfig.Port
		} else {
			rabbitConfig.Host = "5672"
		}

		if strings.TrimSpace(cfg.RabbitConfig.QueueName) != "" {
			rabbitConfig.QueueName = cfg.RabbitConfig.QueueName
		} else {
			rabbitConfig.QueueName = "WorkerQueue"
		}

		/*
			if cfg.RabbitConfig.ResultsExpireIn > 0 {
				rabbitConfig.ResultsExpireIn = cfg.RabbitConfig.ResultsExpireIn
			} else {
				rabbitConfig.ResultsExpireIn = 1
			}
		*/

		if strings.TrimSpace(cfg.RabbitConfig.UserName) != "" {
			rabbitConfig.UserName = cfg.RabbitConfig.UserName
		} else {
			rabbitConfig.UserName = "guest"
		}
		return rabbitConfig
	}
	return nil
}

func buildConfig(cfg Config) Config {
	config := Config{}
	config.RabbitConfig = buildRabbitConfig(cfg)
	config.LocalMode = cfg.LocalMode

	return config
}
