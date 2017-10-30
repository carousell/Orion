package orion

import (
	"errors"
	"reflect"
)

func decoder(in interface{}) error {
	if in == nil {
		return errors.New("No input object!")
	}
	t := reflect.TypeOf(in)
	if t.Kind() != reflect.Struct {
		return errors.New("decoder can only deserialize to structs, can not convert " + t.String() + " of kind " + t.Kind().String())
	}
	return nil
}
