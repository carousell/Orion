package orion

import (
	"errors"
	"log"
	"os"
	"reflect"
)

func decoder(in interface{}) error {
	if in == nil {
		return errors.New("No input object")
	}
	t := reflect.TypeOf(in)
	if t.Kind() != reflect.Struct {
		return errors.New("decoder can only deserialize to structs, can not convert " + t.String() + " of kind " + t.Kind().String())
	}
	return nil
}

func getHostname() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	log.Println("HOST", host)
	return host
}
