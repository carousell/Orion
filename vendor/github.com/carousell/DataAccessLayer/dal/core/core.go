//Package core provides the core implementation of dal.DataAcessLayer that works across data stores
package core

import (
	"context"
	"reflect"

	"github.com/carousell/DataAccessLayer/dal"
	"github.com/carousell/DataAccessLayer/dal/cassandra"
	"github.com/carousell/DataAccessLayer/dal/es"
	"github.com/pkg/errors"
	elastic "gopkg.in/olivere/elastic.v5"
)

type coreDAL struct {
	cassandra dal.DataAccessLayer
	elastic   dal.DataAccessLayer
}

func (c *coreDAL) Initialize() error {
	err := c.cassandra.Initialize()
	if err != nil {
		return err
	}
	if c.elastic != nil {
		return c.elastic.Initialize()
	}
	return nil
}

func (c *coreDAL) Close() {
	c.cassandra.Close()
	if c.elastic != nil {
		c.elastic.Close()
	}
}

func (c *coreDAL) Insert(ctx context.Context, record interface{}) error {
	err := c.cassandra.Insert(ctx, record)
	if err == nil && c.elastic != nil {
		return c.elastic.Insert(ctx, record)
	}
	return err
}

func (c *coreDAL) Delete(ctx context.Context, record interface{}) error {
	err := c.cassandra.Delete(ctx, record)
	if err == nil && c.elastic != nil {
		return c.elastic.Delete(ctx, record)
	}
	return err
}

func (c *coreDAL) Upsert(ctx context.Context, record interface{}, omitempty bool) error {
	err := c.cassandra.Upsert(ctx, record, omitempty)
	if err == nil && c.elastic != nil {
		return c.elastic.Upsert(ctx, record, omitempty)
	}
	return err
}

func (c *coreDAL) ReadPrimary(ctx context.Context, key interface{}) error {
	if reflect.TypeOf(key).Kind() != reflect.Ptr {
		return dal.ErrorPointerNeeded
	}
	return c.cassandra.ReadPrimary(ctx, key)
}

func (c *coreDAL) Find(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKeys []elastic.Sorter) ([]interface{}, error) {
	if c.elastic != nil {
		return c.elastic.Find(ctx, record, query, offset, size, sortKeys)
	}
	return make([]interface{}, 0), errors.New("Elasticsearch not initialized")
}

func (c *coreDAL) FindConsistent(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKeys []elastic.Sorter) ([]interface{}, error) {
	result, err := c.Find(ctx, record, query, offset, size, sortKeys)
	if err == nil {
		for i := range result {
			c.ReadPrimary(ctx, &result[i])
		}
	}
	return result, err
}

func (c *coreDAL) Count(ctx context.Context, record interface{}, query elastic.Query) (int64, error) {
	if c.elastic != nil {
		return c.elastic.Count(ctx, record, query)
	}
	return 0, errors.New("Elasticsearch not initialized")
}

//NewClient initializes a new client implementation of dal.DataAccessLayer.
func NewClient(configs ...interface{}) (dal.DataAccessLayer, error) {
	d := new(coreDAL)
	for _, config := range configs {
		if cas, ok := config.(cassandra.Config); ok {
			var err error
			d.cassandra, err = cassandra.NewClient(cas)
			if err != nil {
				return nil, err
			}
		} else if e, ok := config.(es.Config); ok {
			var err error
			d.elastic, err = es.NewClient(e)
			if err != nil {
				return nil, err
			}
		}
	}
	return d, nil
}
