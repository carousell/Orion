package orion

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/carousell/Orion/orion/handlers"
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

func RegisterEncoder(svr Server, serviceName, method, httpMethod, path string, encoder handlers.Encoder) {
	fmt.Println("registering encoder for", method, path)
	if e, ok := svr.(handlers.Encodeable); ok {
		e.AddEncoder(serviceName, method, httpMethod, path, encoder)
	}
}
