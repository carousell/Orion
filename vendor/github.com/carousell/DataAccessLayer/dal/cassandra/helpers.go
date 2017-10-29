package cassandra

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/carousell/DataAccessLayer/dal/marshaller"
	"github.com/carousell/DataAccessLayer/dal/marshaller/helper"
)

//CreateCassandraTableWithPrefix is a helper function that returns DataAccessLayer's view of how a structure is represented in cassandra table.
//
//This function can also be used to generate a cassandra table from a provided structure when a prefix is being used
func CreateCassandraTableWithPrefix(record interface{}, prefix string) string {
	table := prefix + helper.GetTableName(record)
	types, _ := helper.InferFields(reflect.ValueOf(record))
	idMap, _ := helper.GetStringIdField(types)

	query := "CREATE TABLE %s (%s, PRIMARY KEY(%s))"
	fields := make([]string, 0, len(types))
	for k, v := range types {
		fields = append(fields, k+" "+getCassandraMapping(v.Value.Type()))
	}
	ids := make([]string, 0)
	for k := range idMap {
		ids = append(ids, k)
	}
	sort.Strings(ids)
	sort.Strings(fields)
	fieldStr := strings.Join(fields, ", ")
	primary := strings.Join(ids, ",")
	return fmt.Sprintf(query, table, fieldStr, primary)
}

//CreateCassandraTable is a helper function that returns DataAccessLayer's view of how a structure is represented in cassandra table.
//
//This function can also be used to generate a cassandra table from a provided structure.
func CreateCassandraTable(record interface{}) string {
	return CreateCassandraTableWithPrefix(record, "")
}

func getCassandraMapping(t reflect.Type) string {
	if t.ConvertibleTo(reflect.TypeOf(marshaller.NullTime{})) {
		return "timestamp"
	} else if t.ConvertibleTo(reflect.TypeOf(marshaller.NullInt64{})) {
		return "bigint"
	} else if t.ConvertibleTo(reflect.TypeOf(marshaller.NullFloat64{})) {
		return "double"
	} else if t.ConvertibleTo(reflect.TypeOf(marshaller.NullBool{})) {
		return "boolean"
	} else if t.ConvertibleTo(reflect.TypeOf(marshaller.NullString{})) {
		return "text"
	}
	return "text"
}
