package es

import (
	"errors"
	"reflect"
	"sort"

	"github.com/carousell/DataAccessLayer/dal/marshaller"
	"github.com/carousell/DataAccessLayer/dal/marshaller/helper"
)

func getTypeName(record interface{}, def string) string {
	if r, ok := record.(ESType); ok {
		return r.GetEsType()
	}
	return def
}

func getIdentifier(record interface{}) (string, error) {
	if r, ok := record.(Identifier); ok {
		return r.GetIdentifier()
	}
	// fetch all fields
	types, err := helper.InferFields(reflect.ValueOf(record))
	if err != nil {
		return "", err
	}

	// check for id field
	idMap, err := helper.GetStringIdField(types)
	if err != nil {
		return "", err
	}

	// build id from primary keys
	id := ""
	keys := make([]string, 0, len(idMap))
	for k, _ := range idMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v, _ := idMap[k]
		var val string = ""
		if i, ok := v.(marshaller.NullString); ok {
			val = i.String
		} else if i, ok := v.(string); ok {
			val = i
		} else {
			return id, errors.New(k + "should be either NullString or string")
		}
		if len(id) == 0 {
			id = val
		} else {
			id = id + ":" + val
		}
	}
	if id == "" {
		return id, errors.New("could not build id for elasticsearch")
	}
	return id, nil
}
