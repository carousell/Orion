package main

import (
	"os"
	"strings"
)

func getHostname() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	logger.Log("HOST", host)
	return host
}

func getServiceLower() string {
	return strings.ToLower(serviceName)
}
