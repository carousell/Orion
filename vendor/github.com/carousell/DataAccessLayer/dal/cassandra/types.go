package cassandra

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
)

var (
	//ErrorInInitialization is thrown when there is an error during initialization
	ErrorInInitialization = errors.New("Error during initialization")
)

type cassandraDAL struct {
	Keyspace                  string
	CassandraHosts            []string
	CassandraConnectTimeout   time.Duration
	CassandraOperationTimeout time.Duration
	cassandraSession          *gocql.Session
	parsedConsistency         gocql.Consistency
	numConns                  int
	logsEnabled               bool
	prefix                    string
}

// Config is the configuration for cassandra that DAL will use to communicate with cassandra
type Config struct {
	//Keyspace should be the cassandra keysapce to use
	Keyspace string
	//CassandraHosts are the hosts that DAL will connect to
	CassandraHosts []string
	//CassandraConsistency is the consistency level for all cassandra calls (ideally this should be set to 'LOCAL_QUORUM')
	CassandraConsistency string
	//CassandraConnectTimeout is the time initial connection to cassandra will wait before timing out
	CassandraConnectTimeout time.Duration
	//CassandraOperationTimeout is the time each operation will wait before timing out
	CassandraOperationTimeout time.Duration
	//NumConns is the number of connections that are maintained per cassandra hosts
	NumConns int
	//EnableLogs enables logging of each cassandra query
	EnableLogs bool
	//Prefix is the prefix to be used for table names
	Prefix string
}
