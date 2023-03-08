package utils

import (
	"context"
	"os"

	"github.com/carousell/Orion/v2/utils/log"
)

//GetHostname fetches the hostname of the system
func GetHostname() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	log.Info(context.Background(), "HOST", host)
	return host
}
