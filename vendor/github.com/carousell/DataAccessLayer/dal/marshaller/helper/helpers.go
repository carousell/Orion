//Package helper contains helper function used by various DataAccessLayer implementations.
package helper

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/carousell/DataAccessLayer/dal"
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
)

const (
	VALID_FIELD_NAME string = "Valid"
)

var (
	types = []reflect.Type{
		reflect.TypeOf((*driver.Valuer)(nil)),
		reflect.TypeOf((*sql.Scanner)(nil)),
		reflect.TypeOf((*gocql.Marshaler)(nil)),
		reflect.TypeOf((*gocql.Unmarshaler)(nil)),
	}
)

type ValueInfo struct {
	Value   reflect.Value
	Pos     int
	Primary bool
}

func IsMarshaller(v reflect.Value) bool {
	v = reflect.Indirect(v)
	t := v.Type()
	for _, typ := range types {
		//fmt.Println("comparing", t.String(), typ.Elem().String())
		if !t.Implements(typ.Elem()) {
			if !reflect.PtrTo(t).Implements(typ.Elem()) {
				//fmt.Println(t.String(), "does not implement", typ.Elem().String())
				return false
			}
		}
	}
	return true
}

func CheckValid(v reflect.Value) bool {
	v = reflect.Indirect(v)
	if IsMarshaller(v) {
		// check for validity
		t := v.Type()
		if _, ok := t.FieldByName(VALID_FIELD_NAME); ok {
			//fmt.Println("checking if value is valid ", v)
			return v.FieldByName(VALID_FIELD_NAME).Bool()
		}
	}
	//fmt.Println("value is not marshaller", v)
	return false
}

func InferFields(v reflect.Value) (map[string]ValueInfo, error) {
	v = reflect.Indirect(v)
	t := v.Type()
	types := make(map[string]ValueInfo)
	numFields := v.NumField()
	if numFields <= 0 || numFields > dal.MAX_STRUCT_FIELDS {
		return types, errors.New("Invalid number of fields in record")
	}
	//log.Println("type is", t.String())

	for i := 0; i < numFields; i++ {
		if t.Field(i).Anonymous {
			continue
		}

		tag := t.Field(i).Tag.Get(dal.TAG_NAME)
		name := ""
		options := ""
		if tag != "" {
			opts := strings.SplitN(tag, ",", 2)
			name = opts[0]
			if len(opts) > 1 {
				options = opts[1]
			}
		} else {
			name = strings.ToLower(t.Field(i).Name)
		}

		if strings.Contains(options, dal.IGNORE_TAG_NAME) {
			continue
		}

		types[name] = ValueInfo{
			Value:   v.Field(i),
			Pos:     i,
			Primary: strings.Contains(options, dal.PRIMARY_TAG_NAME),
		}
	}
	return types, nil
}

//tablerAlt (Deprecated) allows use of custom table/index name
type tablerAlt interface {
	GetTableNameAlt() string
}

func GetTableName(record interface{}) string {
	if r, ok := record.(dal.Tabler); ok {
		if strings.TrimSpace(r.GetTableName()) != "" {
			return r.GetTableName()
		}
	}
	//TODO remove this
	if r, ok := record.(tablerAlt); ok {
		if strings.TrimSpace(r.GetTableNameAlt()) != "" {
			return r.GetTableNameAlt()
		}
	}
	// default name of table
	v := reflect.ValueOf(record)
	v = reflect.Indirect(v)
	return strings.ToLower(v.Type().Name())
}

func BuildValueInfoMap(value reflect.Value, types map[string]ValueInfo) (map[string]ValueInfo, error) {
	newMap := make(map[string]ValueInfo)

	value = reflect.Indirect(value)
	numFields := value.NumField()
	for k, v := range types {
		if v.Pos <= numFields {
			newMap[k] = ValueInfo{Value: value.Field(v.Pos), Pos: v.Pos}
		} else {
			return newMap, errors.New("value has less fields than map")
		}
	}

	return newMap, nil
}

func sortTypesKeys(types map[string]ValueInfo) []string {
	keys := make([]string, 0, len(types))
	for k, _ := range types {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func BuildKeyValue(types map[string]ValueInfo) ([]string, []interface{}) {
	numFields := len(types)
	keys := make([]string, 0, numFields)
	values := make([]interface{}, 0, numFields)
	for _, k := range sortTypesKeys(types) {
		keys = append(keys, k)
		values = append(values, types[k].Value.Interface())
	}
	return keys, values
}

func BuildKeyValueForScanner(types map[string]ValueInfo) ([]string, []interface{}) {
	values := make([]interface{}, 0, len(types))
	keys := make([]string, 0, len(types))
	for _, k := range sortTypesKeys(types) {
		v := types[k]
		if v.Value.CanAddr() {
			keys = append(keys, k)
			values = append(values, v.Value.Addr().Interface())
		} else {
			log.Println("cannot addr", k, v)
		}
	}
	return keys, values
}

func BuildKeyValueForScannerFromReferance(value reflect.Value, types map[string]ValueInfo) ([]string, []interface{}, error) {
	value = reflect.Indirect(value)
	numFields := value.NumField()

	values := make([]interface{}, 0, len(types))
	keys := make([]string, 0, len(types))
	for _, k := range sortTypesKeys(types) {
		v := types[k]
		if v.Pos <= numFields {
			fieldValue := value.Field(v.Pos)
			if fieldValue.CanAddr() {
				keys = append(keys, k)
				values = append(values, fieldValue.Addr().Interface())
			} else {
				log.Println("cannot addr", k, fieldValue)
			}
		} else {
			return keys, values, errors.New("value has less fields than map")
		}
	}
	return keys, values, nil
}

func BuildOrderedKeyValueForScannerFromReferance(value reflect.Value, types map[string]ValueInfo, keyOrder []string) ([]string, []interface{}, error) {
	value = reflect.Indirect(value)
	numFields := value.NumField()

	values := make([]interface{}, 0, len(types))
	keys := make([]string, 0, len(types))
	for _, k := range keyOrder {
		v, ok := types[k]
		if ok && v.Pos <= numFields {
			fieldValue := value.Field(v.Pos)
			if fieldValue.CanAddr() {
				keys = append(keys, k)
				values = append(values, fieldValue.Addr().Interface())
			} else {
				log.Println("cannot addr", k, fieldValue)
			}
		} else {
			return keys, values, errors.New("value has less fields than map")
		}
	}
	return keys, values, nil
}

func GetStringIdField(types map[string]ValueInfo) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	for k, v := range types {
		if v.Primary {
			val := reflect.Indirect(v.Value)
			m[k] = val.Interface()
		}
	}
	if len(m) == 0 {
		return m, errors.New("id field not found")
	}
	return m, nil
}

func InspectStructV(val reflect.Value) {
	if val.Kind() == reflect.Interface && !val.IsNil() {
		elm := val.Elem()
		if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
			val = elm
		}
	}

	val = reflect.Indirect(val)

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		address := "not-addressable"

		if val.Type().Field(i).Anonymous {
			continue
		}

		if valueField.Kind() == reflect.Interface && !valueField.IsNil() {
			elm := valueField.Elem()
			if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
				valueField = elm
			}
		}

		if valueField.Kind() == reflect.Ptr {
			valueField = valueField.Elem()

		}
		if valueField.CanAddr() {
			address = fmt.Sprintf("0x%X", valueField.Addr().Pointer())
		}

		fmt.Printf("Field Name: %s,\t Field Value: %v,\t Address: %v\t, Field type: %v\t, Field kind: %v\n", typeField.Name,
			valueField.Interface(), address, typeField.Type, valueField.Kind())

		if valueField.Kind() == reflect.Struct {
			InspectStructV(valueField)
		}
	}
}
