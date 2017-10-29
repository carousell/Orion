package dal

import (
	"context"

	"github.com/pkg/errors"
	elastic "gopkg.in/olivere/elastic.v5"
)

var (
	// ErrorPointerNeeded when a value is not passed as a reference
	ErrorPointerNeeded = errors.New("need pointer type")
	//ErrorNotImplemented when a feature is invoked that is not supported by the current set on datastores
	ErrorNotImplemented = errors.New("Not Implemented")
	//ErrorNotfound  when the requested item was not found
	ErrorNotFound = errors.New("Not found")
)

// constant values used throughout DAL
const (
	MAX_STRUCT_FIELDS int    = 50        // max number of fields supported in a struct
	TAG_NAME          string = "json"    // name of the tag that DAL reads for info
	PRIMARY_TAG_NAME  string = "primary" // tag name to mark a field as primary key
	IGNORE_TAG_NAME   string = "ignore"  // tag name to mark a field as ignore
)

// Tabler allows use of custom table/index names (should be implemented as non pointer reciever)
type Tabler interface {
	GetTableName() string
}

// DataAccessLayer is the interface for all interactions with DAL
type DataAccessLayer interface {
	//Initialize initializes all the connections that are used in DAL (normally this is called by DAL's NewClient)
	Initialize() error
	//Closes all underlying transports
	Close()
	//Insert inserts data into the initialized storage backends and overrides any existing data
	Insert(ctx context.Context, record interface{}) error
	//Delete deletes the given record from storage backends, you only need to populate the primary keys for a given struct
	Delete(ctx context.Context, record interface{}) error
	//Upsert updates the given record in storage backends, you need to populate the primary keys and any other keys that needs to be updated
	Upsert(ctx context.Context, record interface{}, omitempty bool) error
	//ReadPrimary reads data from the primary data store (currently cassandra), you only need to populate the primary keys for a given struct
	ReadPrimary(ctx context.Context, key interface{}) error
	//Find searches for data provided in the query and is sorted based on keys provided in sortKeys
	Find(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKeys []elastic.Sorter) ([]interface{}, error)
	//FindConsistent searches for data using find and then enriches that data from primary datastore
	FindConsistent(ctx context.Context, record interface{}, query elastic.Query, offset, size int, sortKeys []elastic.Sorter) ([]interface{}, error)
	//Count provides a count of records found in the query
	Count(ctx context.Context, record interface{}, query elastic.Query) (int64, error)
}
