//Package es provides implementation of dal.DataAccessLayer for ElasticSearch.
//
//Note: Clients wanting to use DataAcessLayer please use the core package.
package es

import (
	"context"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/carousell/DataAccessLayer/dal"
	"github.com/carousell/DataAccessLayer/dal/marshaller/helper"
	"github.com/carousell/go-utils/utils/spanutils"
	"github.com/pkg/errors"

	elastic "gopkg.in/olivere/elastic.v5"
)

//NewClient takes a es.Config and returns an implementation of dal.DataAccessLayer talking to that elastic search cluster.
//
//Note: Clients wanting to use DataAcessLayer please use the core package
func NewClient(config Config) (ES, error) {
	e := new(elasticsearch)
	e.url = config.Url
	e.prefix = config.Prefix
	e.fakeContext = config.FakeContext
	e.createIndex = false
	e.debugLogs = config.DebugLog
	e.username = config.UserName
	e.password = config.Password
	return e, e.Initialize()
}

type elasticsearch struct {
	client      *elastic.Client
	url         string
	prefix      string
	createIndex bool
	fakeContext bool
	debugLogs   bool
	username    string
	password    string
}

func (e *elasticsearch) Initialize() error {
	connOpts := []elastic.ClientOptionFunc{
		elastic.SetURL(e.url),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
	}
	if e.username != "" || e.password != "" {
		connOpts = append(connOpts, elastic.SetBasicAuth(e.username, e.password))
	}
	if e.debugLogs {
		connOpts = append(connOpts, elastic.SetInfoLog(log.New(os.Stdout, "ELASTIC ", log.LstdFlags)))
		connOpts = append(connOpts, elastic.SetTraceLog(log.New(os.Stdout, "ELASTIC ", log.LstdFlags)))
	}
	// Create a client
	client, err := elastic.NewClient(
		connOpts...,
	)
	if err != nil {
		// Handle error
		return err
	}
	e.client = client
	return nil
}

func (e *elasticsearch) Close() {
	e.client.Stop()
}

func (e *elasticsearch) getIndexName(record interface{}) string {
	return e.prefix + helper.GetTableName(record)
}

func (e *elasticsearch) checkAndCreateIndex(ctx context.Context, indexName string) error {
	if !e.createIndex {
		return nil
	}
	exists, err := e.client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return err
	}
	if !exists {
		_, err := e.client.CreateIndex(indexName).Do(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *elasticsearch) Insert(ctx context.Context, record interface{}) error {
	name := "DALElasticSearchInsert"
	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALElasticSearch")
	defer span.Finish()

	indexName := e.getIndexName(record)

	id, err := getIdentifier(record)
	if err != nil {
		return err
	}

	if e.createIndex {
		err := e.checkAndCreateIndex(ctx, indexName)
		if err != nil {
			return err
		}
	}
	span.SetTag("index", indexName)
	span.SetTag("id", id)

	err = hystrix.Do(name, func() error {
		var err error
		q := e.client.Index().
			Index(indexName).
			Type(getTypeName(record, indexName)).
			Id(id).
			BodyJson(record)
		//Routing() TODO add support for routing
		if e.fakeContext {
			_, err = q.Do(context.Background())
		} else {
			_, err = q.Do(ctx)
		}
		return err
	}, nil)

	if err != nil {
		span.SetTag("error", err.Error())
		span.SetTag("failure", "true")
	}
	return err
}

func (e *elasticsearch) FindWithHandler(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKeys []elastic.Sorter, handler HitHandler, explain bool) ([]interface{}, error) {
	name := "DALElasticSearchFind"

	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALElasticSearch")
	defer span.Finish()

	indexName := e.getIndexName(record)
	span.SetTag("index", indexName)

	if offset < 0 {
		offset = 0
	}

	if size > 1000 {
		size = 1000
		log.Println("DAL: to more then 1000 items please use query API")
	} else if size < 1 {
		size = 10
	}

	// fetch all fields
	types, err := helper.InferFields(reflect.ValueOf(record))
	if err != nil {
		return make([]interface{}, 0), err
	}

	fields := make([]string, 0, len(types))
	for k := range types {
		fields = append(fields, k)
	}

	sourceContext := elastic.NewFetchSourceContext(true)
	sourceContext.Include(fields...)

	q := e.client.Search(indexName).Query(query).From(offset).Size(size).Type(getTypeName(record, indexName)).FetchSourceContext(sourceContext).Explain(explain)
	if len(sortKeys) > 0 {
		q = q.SortBy(sortKeys...)
	}
	var res *elastic.SearchResult

	err = hystrix.Do(name, func() error {
		var err error
		if e.fakeContext {
			res, err = q.Do(context.Background())
		} else {
			res, err = q.Do(ctx)
		}
		return err
	}, nil)
	if err != nil {
		span.SetTag("error", err.Error())
		span.SetTag("failure", "true")
		return nil, err
	}
	span.SetTag("took", time.Duration(res.TookInMillis)*time.Millisecond)
	span.SetTag("hits", res.TotalHits())
	return handler(res, record)

}
func (e *elasticsearch) Find(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKey []elastic.Sorter) ([]interface{}, error) {
	return e.FindWithHandler(ctx, record, query, offset, size, sortKey, defaultHitHandler, false)
}

func (e *elasticsearch) FindConsistent(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKeys []elastic.Sorter) ([]interface{}, error) {
	return make([]interface{}, 0), errors.WithStack(dal.ErrorNotImplemented)
}

func (e *elasticsearch) Delete(ctx context.Context, record interface{}) error {
	name := "DALElasticSearchDelete"
	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALElasticSearch")
	defer span.Finish()

	indexName := e.getIndexName(record)

	id, err := getIdentifier(record)
	if err != nil {
		return err
	}

	span.SetTag("index", indexName)
	span.SetTag("id", id)

	hystrix.Do(name, func() error {
		q := e.client.Delete().
			Index(indexName).
			Type(getTypeName(record, indexName)).
			Id(id)
		//Routing() TODO add support for routing
		if e.fakeContext {
			_, err = q.Do(context.Background())
		} else {
			_, err = q.Do(ctx)
		}
		if elastic.IsNotFound(err) {
			span.SetTag("es-error", err.Error())
			err = dal.ErrorNotFound
			return nil
		}
		return err
	}, nil)

	if err != nil {
		span.SetTag("error", err.Error())
		span.SetTag("failure", "true")
	}
	return err
}

func (e *elasticsearch) Upsert(ctx context.Context, record interface{}, omitempty bool) error {
	name := "DALElasticSearchUpsert"
	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALElasticSearch")
	defer span.Finish()

	indexName := e.getIndexName(record)

	// fetch all fields
	types, err := helper.InferFields(reflect.ValueOf(record))
	if err != nil {
		return err
	}
	id, err := getIdentifier(record)
	if err != nil {
		return err
	}
	span.SetTag("index", indexName)
	span.SetTag("id", id)

	doc := make(map[string]interface{})
	for k, v := range types {
		if omitempty == true {
			if !helper.CheckValid(v.Value) {
				continue
			}
		}
		doc[k] = v.Value.Interface()
	}

	err = hystrix.Do(name, func() error {
		var err error
		q := e.client.Update().
			Index(indexName).Type(getTypeName(record, indexName)).
			Doc(doc).Id(id).DocAsUpsert(true)
		//Routing() TODO add support for routing
		if e.fakeContext {
			_, err = q.Do(context.Background())
		} else {
			_, err = q.Do(ctx)
		}
		return err
	}, nil)

	if err != nil {
		span.SetTag("error", err.Error())
		span.SetTag("failure", "true")
	}
	return err
}

func (e *elasticsearch) ReadPrimary(ctx context.Context, key interface{}) error {
	return errors.WithStack(dal.ErrorNotImplemented)
}

func (e *elasticsearch) Count(ctx context.Context, record interface{}, query elastic.Query) (int64, error) {
	name := "DALElasticSearchCount"

	// zipkin span
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALElasticsearch")
	defer span.Finish()

	indexName := e.getIndexName(record)
	span.SetTag("index", indexName)

	q := e.client.Count(indexName).Query(query)
	var res int64
	err := hystrix.Do(name, func() error {
		var err error
		if e.fakeContext {
			res, err = q.Do(context.Background())
		} else {
			res, err = q.Do(ctx)
		}
		return err
	}, nil)

	if err != nil {
		span.SetTag("error", err.Error())
		span.SetTag("failure", "true")
		return 0, err
	}
	span.SetTag("count", res)
	return res, err
}
func (e *elasticsearch) BulkIndex(ctx context.Context, indexables []Indexable) (BulkIndexResponse, error) {
	name := "DALElasticBulkIndex"
	r := BulkIndexResponse{}
	requsets := make([]elastic.BulkableRequest, 0, len(indexables))
	for _, indexable := range indexables {
		b, er := indexable.MakeBulkable()
		if er != nil {
			return r, er
		}
		requsets = append(requsets, b...)
	}
	span, ctx := spanutils.NewDatastoreSpan(ctx, name, "DALELasticserch")
	defer span.Finish()
	err := hystrix.Do(name, func() error {
		response, er := e.client.Bulk().Add(requsets...).Do(ctx)
		if er != nil {
			return er
		}
		r.DeleteCount = len(response.Deleted())
		r.IndexCount = len(response.Indexed())
		r.UpdateCount = len(response.Updated())
		for _, f := range response.Failed() {
			reason := ""
			if f.Error != nil {
				reason = f.Error.Reason
			}
			r.Failures = append(r.Failures, IndexFailures{f.Id, f.Index, reason})
		}
		return er
	}, nil)
	if err != nil {
		span.SetTag("error", err.Error())
		span.SetTag("failure", "true")
		return r, err
	}
	span.SetTag("Result", r)
	return r, nil
}

func defaultHitHandler(res *elastic.SearchResult, clientType interface{}) ([]interface{}, error) {
	return res.Each(reflect.TypeOf(clientType)), nil
}
