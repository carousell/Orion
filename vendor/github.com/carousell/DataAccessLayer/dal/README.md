# dal
`import "."`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Package dal provides implementation of DataAcessLayer, it aims to simplify and standardize access to data stores within Carousell.

Source code for DAL can be found at <a href="https://github.com/carousell/DataAccessLayer">https://github.com/carousell/DataAccessLayer</a>

Note: Use package 'core' to get the default implementation of DataAccessLayer (this is what you need in most cases),
in case you need to provide your own implementation you can use package 'cassandra' and 'es' to get their implementation
and override/extend functionality

Core implementation of DataAccessLayer revolves around 'cassandra' and 'elasticsearch', you should initialize both to be able to utilize the full functionality,

Features supported with 'cassandra'

	Insert
	Delete
	Upsert
	ReadPrimary

features supported with 'elasticsearch'

	Find
	FindConsistent
	Count

### Defining Structs
Defining structures for DAL is easy, have a look at the structure below

	type Group struct {
		ID          marshaller.NullString `json:"id,ignore"` // this field is ignored by Cassandra DAL
		Name        marshaller.NullString `json:"name"`
		Code        marshaller.NullString `json:"code"`
		CountryCode marshaller.NullString `json:"country_code"`
		Description marshaller.NullString `json:"description"`
		URL         marshaller.NullString `json:"url"`
		Created     marshaller.NullTime   `json:"created"`
		Modified    marshaller.NullTime   `json:"modified"`
		UUID        marshaller.NullString `json:"uuid,primary"` // this is the primary key
	}

As you can see we are using the 'marshaller' package, 'marshaller' allows us to define fields that are usable across SQL, Cassandra and ElasticSearch, which allows us to use the same structure for all of them.

We use tags to identify different fields, in the example above we use 'ignore' tag to tell cassandra not to include this field in our schema, also we use 'primary' tag to tell cassandra that this field is our primary field

Note: DAL does not support nested structures and assumes a flat hierarchy

### Table Names
DAL assumes the name of structure to be same as name of table/index, you should implement 'dal.Tabler' as non pointer receiver if you want custom table/index names

In the above defined groups DAL will identify table/index name as 'group', but if we implement 'dal.Tabler' as follows

	func (g Group) GetTableName() string {
		return "groups_data"
	}

DAL will start using table/index name as 'groups_data'

Note: DAL does not create tables in cassandra or indexes in elasticsearch, you need to create those first.

### Custom ES Types
DAL assumes the name of Table/Index as type in elasticsearch index, you should implement 'es.ESType' as a non pointer receiver if you want a custom type for the structure

In the above defined groups DAL will identify type as 'group', but if we implement 'es.ESType' as follows

	func (g Group) GetEsType() string {
		return "groupsInfo"
	}

DAL will start using type as 'groupsInfo'

Note: ESType defaults to index name if not defined, so if index name is overridden ESType will start using index name as type.
Priority order ESType <- Index Name <- Default struct name

### Creating Tables and Indexes
To create table in 'cassandra' we can use the helper functions provided by cassandra

Lets take a structure we want to represent in DAL

	type BlockedUser struct {
		ID          marshaller.NullString `json:"id,ignore"`
		UUID        marshaller.NullString `json:"uuid,ignore"`
		TimeBlocked marshaller.NullTime   `json:"time_blocked"`
		GroupID     marshaller.NullString `json:"group_id"`
		GroupUUID   marshaller.NullString `json:"group_uuid,primary"`
		UserID      marshaller.NullString `json:"user_id,primary"`
	}

here we have two fields marked as primary GroupUUID and UserID, all cassandra lookups will happen on these keys

To generate a table mapping out of this we can use the following code

	fmt.Println(cassandra.CreateCassandraTable(BlockedUser{}))

which will give us the following output

	CREATE TABLE blockeduser (group_id text, group_uuid text, time_blocked timestamp, user_id text, PRIMARY KEY(group_uuid,user_id))

You can use this to create the table in cassandra

### Initialization
Using DAL is easy, just use the default DAL implementation provided in the 'core' package

	esConfig := es.Config{}
	esConfig.Url = "<a href="http://10.140.0.100:9200">http://10.140.0.100:9200</a>"
	esConfig.Prefix = "groups_"
	
	casConfig := cassandra.Config{
		Keyspace:                  "groups",
		CassandraHosts:            []string{"10.140.0.101"},
		CassandraConsistency:      "LOCAL_QUORUM",
		CassandraConnectTimeout:   5 * time.Second,
		CassandraOperationTimeout: 200 * time.Millisecond,
	}
	dalClient, err := core.NewClient(casConfig, esConfig)
	if err != nil {
		log.Println(err)
	}

Now you can use 'dalClient' to access DataAccessLayer implementation, this core implementation automatically calls appropriate datastore based on the operation performed

### Inserting Data
Inserting data is easy, just create the record and call 'Insert' function

	g := Group{}
	
	// fill in data
	t := time.Now() // initialize time so that its same throughout
	g.Created.Scan(t)
	g.Modified.Scan(t)
	g.Code.Scan("coffee_sg")
	g.CountryCode.Scan("SG")
	g.Description.Scan("A group for all coffee lovers")
	g.Name.Scan("Coffee SG")
	g.URL.Scan("coffee-sg")
	g.UUID.Scan("be997e92-c6be-4083-99a5-f01feacf5950")
	
	// write to dal
	err := dalClient.Insert(ctx, g)
	if err != nil {
		log.Println(err)
	}

All the fields that are populated will be written to
Cassandra and ElasticSearch

Note: DAL does not create tables in cassandra or indexes in elasticsearch, you need to create those first

### Reading a Row
Reading data is easy, we need to provide all primary keys for the struct and then call ReadPrimary function

	g := Group{}
	g.UUID.Scan("be997e92-c6be-4083-99a5-f01feacf5950")
	
	// read from dal
	err := dalClient.ReadPrimary(ctx, &g)
	if err != nil {
		log.Println(err)
	}

Data will be populated in the same structure 'g' in the above example

### Searching for Data
DataAccessLayer exposes complex ElasticSearch queries using the elastic library, make sure your code imports the following

	import elastic "gopkg.in/olivere/elastic.v5"

We can create a query as follows

	q := elastic.NewBoolQuery()
	q.Filter(
		elastic.NewTermQuery("country_code.keyword", "SG")
	)
	m := elastic.NewMultiMatchQuery("coffee", "name", "code")
	m.Type("phrase_prefix")
	m.Operator("and")
	q.Must(m)
	
	s := elastic.NewFieldSort("modified").Desc()
	
	fmt.Println(dalClient.Find(ctx, g, q, 0, 10, []elastic.Sorter{s}))

In the above code we search for all groups containing "coffee" in their name or code and having country_code=SG .

## <a name="pkg-imports">Imported Packages</a>

- github.com/pkg/errors
- gopkg.in/olivere/elastic.v5

## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [type DataAccessLayer](#DataAccessLayer)
* [type Tabler](#Tabler)

#### <a name="pkg-files">Package files</a>
[doc.go](./doc.go) [types.go](./types.go) 

## <a name="pkg-constants">Constants</a>
``` go
const (
    MAX_STRUCT_FIELDS int    = 50        // max number of fields supported in a struct
    TAG_NAME          string = "json"    // name of the tag that DAL reads for info
    PRIMARY_TAG_NAME  string = "primary" // tag name to mark a field as primary key
    IGNORE_TAG_NAME   string = "ignore"  // tag name to mark a field as ignore
)
```
constant values used throughout DAL

## <a name="pkg-variables">Variables</a>
``` go
var (
    // ErrorPointerNeeded when a value is not passed as a reference
    ErrorPointerNeeded = errors.New("need pointer type")
    //ErrorNotImplemented when a feature is invoked that is not supported by the current set on datastores
    ErrorNotImplemented = errors.New("Not Implemented")
    //ErrorNotfound  when the requested item was not found
    ErrorNotFound = errors.New("Not found")
)
```

## <a name="DataAccessLayer">type</a> [DataAccessLayer](./types.go#L33-L52)
``` go
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
```
DataAccessLayer is the interface for all interactions with DAL

## <a name="Tabler">type</a> [Tabler](./types.go#L28-L30)
``` go
type Tabler interface {
    GetTableName() string
}
```
Tabler allows use of custom table/index names (should be implemented as non pointer reciever)

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)