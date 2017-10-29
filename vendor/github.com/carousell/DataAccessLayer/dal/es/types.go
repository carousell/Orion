package es

import (
	"context"

	"github.com/carousell/DataAccessLayer/dal"

	elastic "gopkg.in/olivere/elastic.v5"
)

// Config is the config used to initialize ElasticSearch connection
type Config struct {
	// URL of the es host to connect to (ideally this can be a HA Node)
	Url string
	// Prefix is the prefix that will be used for index name and type
	Prefix string
	// FakeContext should be set when we want the request to es to be completed even if the base context passed to DAL interface is cancelled
	FakeContext bool
	// Enable Debug Logs
	DebugLog bool
	// Basic Authentication
	UserName string
	Password string
}

// ESType allows use of custom type in ElasticSearch (should be implemented as non pointer reciever)
type ESType interface {
	GetEsType() string
}

// Identifier allows clients to have custom ids.
type Identifier interface {
	GetIdentifier() (string, error)
}

type ES interface {
	ESQueryLayer
	ESIndexer
	dal.DataAccessLayer
}

type ESQueryLayer interface {
	FindWithHandler(context.Context, interface{}, elastic.Query, int, int, []elastic.Sorter, HitHandler, bool) ([]interface{}, error)
}
type ESIndexer interface {
	BulkIndex(context.Context, []Indexable) (BulkIndexResponse, error)
}
type BulkIndexResponse struct {
	UpdateCount, DeleteCount, IndexCount int
	Failures                             []IndexFailures
}
type IndexFailures struct {
	Id, Index, Reason string
}
type Indexable interface {
	Identifier
	ESType
	MakeBulkable() ([]elastic.BulkableRequest, error)
}

// Allow clients to provide custom code to handle results
type HitHandler func(*elastic.SearchResult, interface{}) ([]interface{}, error)
