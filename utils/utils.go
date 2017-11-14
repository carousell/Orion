package utils

import (
	"log"
	"os"
)

func GetHostname() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	log.Println("HOST", host)
	return host
}
