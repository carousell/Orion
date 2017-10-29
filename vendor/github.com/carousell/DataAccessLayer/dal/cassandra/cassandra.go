//Package cassandra provides implementation of dal.DataAccessLayer for cassandra.
//
//Note: Clients wanting to use DataAcessLayer please use the core package.
package cassandra

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/DataAccessLayer/dal"
	"github.com/carousell/DataAccessLayer/dal/marshaller/helper"
	"github.com/carousell/go-utils/utils/spanutils"
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
)

//NewClient takes a cassandra.Config and returns an implementation of dal.DataAccessLayer talking to that cassandra cluster.
//
//Note: Clients wanting to use DataAcessLayer please use the core package
func NewClient(config Config) (dal.DataAccessLayer, error) {
	client := new(cassandraDAL)
	//TODO Handle defaults
	if len(config.CassandraHosts) < 1 {
		return nil, errors.WithStack(ErrorInInitialization)
	}
	client.CassandraHosts = config.CassandraHosts

	if config.CassandraConnectTimeout == 0 {
		config.CassandraConnectTimeout = time.Second * 5
	}
	client.CassandraConnectTimeout = config.CassandraConnectTimeout

	if config.CassandraOperationTimeout == 0 {
		config.CassandraOperationTimeout = time.Second * 5
	}
	client.CassandraOperationTimeout = config.CassandraOperationTimeout

	if strings.TrimSpace(config.Keyspace) == "" {
		return nil, errors.New("Empty Keyspace")
	}
	client.Keyspace = config.Keyspace

	if config.NumConns == 0 {
		config.NumConns = 4
	}
	client.numConns = config.NumConns

	var err error
	client.parsedConsistency, err = gocql.ParseConsistencyWrapper(config.CassandraConsistency)
	if err != nil {
		return nil, err
	}

	client.logsEnabled = config.EnableLogs
	client.prefix = config.Prefix

	err = client.Initialize()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (d *cassandraDAL) Initialize() error {
	err := d.reinitializeCassandra()
	return err
}

func (d *cassandraDAL) Close() {
	d.cassandraSession.Close()
}

func (d *cassandraDAL) reinitializeCassandra() error {
	if d.cassandraSession != nil {
		d.cassandraSession.Close()
		d.cassandraSession = nil
	}
	cluster := gocql.NewCluster(d.CassandraHosts...)

	cluster.Keyspace = d.Keyspace
	cluster.ConnectTimeout = d.CassandraConnectTimeout
	cluster.Timeout = d.CassandraOperationTimeout
	cluster.Consistency = d.parsedConsistency
	cluster.NumConns = d.numConns

	var err error
	d.cassandraSession, err = cluster.CreateSession()
	return errors.Wrap(err, "Could not initialize cassandra")
}

func (d *cassandraDAL) getTableName(record interface{}) string {
	return d.prefix + helper.GetTableName(record)
}

func (d *cassandraDAL) Insert(ctx context.Context, record interface{}) error {
	name := "DALCassandraInsert"
	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALCassandra")
	defer span.Finish()

	tableName := d.getTableName(record)

	// fetch all fields
	types, err := helper.InferFields(reflect.ValueOf(record))
	if err != nil {
		return err
	}

	// check for id field
	_, err = helper.GetStringIdField(types)
	if err != nil {
		return err
	}

	// build placeholders
	numFields := len(types)
	placeHolders := make([]string, numFields, numFields)
	for i := 0; i < numFields; i++ {
		placeHolders[i] = "?"
	}

	// build keys and values
	keys, values := helper.BuildKeyValue(types)

	queryStr := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		d.Keyspace,
		tableName,
		strings.Join(keys, ", "),
		strings.Join(placeHolders, ", "))
	if d.logsEnabled {
		log.Println("Executing", queryStr, values)
	}

	span.SetQuery(queryStr)

	err = nil
	e := hystrix.Do(name, func() error {
		err = d.cassandraSession.Query(queryStr, values...).Exec()
		if err == gocql.ErrNotFound {
			err = dal.ErrorNotFound
			return nil
		}
		return errors.Wrap(err, "Could not insert data in cassandra")
	}, nil)
	// check for hystrix related error and return it
	if err == nil && e != nil {
		err = e
	}
	if err != nil {
		span.SetTag("error", err.Error())
		span.SetTag("failure", "true")
	}
	return err
}

func (d *cassandraDAL) Upsert(ctx context.Context, record interface{}, omitempty bool) error {
	name := "DALCassandraUpsert"
	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALCassandra")
	defer span.Finish()

	tableName := d.getTableName(record)

	types, err := helper.InferFields(reflect.ValueOf(record))
	if err != nil {
		return err
	}

	idMap, err := helper.GetStringIdField(types)
	if err != nil {
		return err
	}

	numFields := len(types)
	updateFields := make([]string, 0, numFields)
	filteredValues := make([]interface{}, 0, numFields+1)

	for k, v := range types {
		if _, ok := idMap[k]; ok {
			continue
		}

		if omitempty == true {
			if !helper.CheckValid(v.Value) {
				continue
			}
		}

		updateFields = append(updateFields, k+"=?")
		filteredValues = append(filteredValues, v.Value.Interface())
	}

	updateFieldsConcat := strings.Join(updateFields, ", ")

	idFields := make([]string, 0, len(idMap))
	for k, v := range idMap {
		idFields = append(idFields, k+"=?")
		filteredValues = append(filteredValues, v)
	}
	idStr := strings.Join(idFields, " and WHERE ")

	queryStr := fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s",
		d.Keyspace,
		tableName,
		updateFieldsConcat,
		idStr)
	if d.logsEnabled {
		log.Println("Executing ", queryStr, filteredValues)
	}

	span.SetQuery(queryStr)

	err = hystrix.Do(name, func() error {
		return d.cassandraSession.Query(queryStr, filteredValues...).Exec()
	}, nil)

	if err != nil {
		span.SetTag("error", err.Error())
		span.SetTag("failure", "true")
	}
	return err
}

func (d *cassandraDAL) Delete(ctx context.Context, key interface{}) error {
	name := "DALCassandraDelete"
	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALCassandra")
	defer span.Finish()

	tableName := d.getTableName(key)
	types, err := helper.InferFields(reflect.ValueOf(key))
	if err != nil {
		return err
	}

	idMap, err := helper.GetStringIdField(types)
	if err != nil {
		return err
	}

	idFields := make([]string, 0, len(idMap))
	idValues := make([]interface{}, 0, len(idMap))
	for k, v := range idMap {
		idFields = append(idFields, k+"=?")
		idValues = append(idValues, v)
	}
	idStr := strings.Join(idFields, " and ")

	queryStr := fmt.Sprintf("DELETE FROM %s.%s WHERE %s",
		d.Keyspace,
		tableName,
		idStr)
	if d.logsEnabled {
		log.Println("executing", queryStr, idValues)
	}
	span.SetQuery(queryStr)
	span.SetTag("id", idValues)

	var casError error
	e := hystrix.Do(name, func() error {
		casError = d.cassandraSession.Query(queryStr, idValues...).Consistency(d.parsedConsistency).Exec()
		if casError == gocql.ErrNotFound {
			err = dal.ErrorNotFound
			return nil
		}
		return casError
	}, nil)
	if casError == nil && e != nil {
		casError = e
	}
	if casError != nil && casError != dal.ErrorNotFound {
		span.SetTag("error", casError.Error())
		span.SetTag("failure", "true")
	}
	return casError
}

func (d *cassandraDAL) ReadPrimary(ctx context.Context, key interface{}) error {
	if reflect.TypeOf(key).Kind() != reflect.Ptr {
		return dal.ErrorPointerNeeded
	}

	name := "DALCassandraReadPrimary"
	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALCassandra")
	defer span.Finish()

	tableName := d.getTableName(key)
	types, err := helper.InferFields(reflect.ValueOf(key))
	if err != nil {
		return err
	}

	idMap, err := helper.GetStringIdField(types)
	if err != nil {
		return err
	}

	keys, values := helper.BuildKeyValueForScanner(types)
	fields := strings.Join(keys, ", ")

	idFields := make([]string, 0, len(idMap))
	idValues := make([]interface{}, 0, len(idMap))
	for k, v := range idMap {
		idFields = append(idFields, k+"=?")
		idValues = append(idValues, v)
	}
	idStr := strings.Join(idFields, " and ")

	queryStr := fmt.Sprintf("SELECT %s FROM %s.%s WHERE %s",
		fields,
		d.Keyspace,
		tableName,
		idStr)
	if d.logsEnabled {
		log.Println("executing", queryStr, idValues)
	}
	span.SetQuery(queryStr)
	span.SetTag("id", idValues)

	var casError error
	e := hystrix.Do(name, func() error {
		casError = d.cassandraSession.Query(queryStr, idValues...).Consistency(d.parsedConsistency).Scan(values...)
		if casError == gocql.ErrNotFound {
			casError = dal.ErrorNotFound
			return nil
		}
		return casError
	}, nil)
	if casError == nil && e != nil {
		casError = e
	}
	if casError != nil && casError != dal.ErrorNotFound {
		span.SetTag("error", casError.Error())
		span.SetTag("failure", "true")
	}
	return casError
}

func (d *cassandraDAL) Find(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKey []elastic.Sorter) ([]interface{}, error) {
	return make([]interface{}, 0), errors.WithStack(dal.ErrorNotImplemented)
}

func (d *cassandraDAL) FindConsistent(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKeys []elastic.Sorter) ([]interface{}, error) {
	return make([]interface{}, 0), errors.WithStack(dal.ErrorNotImplemented)
}

func (d *cassandraDAL) Count(ctx context.Context, record interface{}, query elastic.Query) (int64, error) {
	return 0, errors.WithStack(dal.ErrorNotImplemented)
}
