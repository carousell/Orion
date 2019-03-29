package main

import (
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/pkg/errors"
	"github.com/carousell/Orion/utils/migrator"
)

// sample only - can be customized based on use-case
type connectConfig struct {
	dbUsername string
	dbPassword string
	dbHostPort string
	dbName     string
}

type clusterData struct {
	name       string
	dbType     string
	sourcePath string
	connConfig connectConfig
}

var clusterMap map[string]clusterData

func getMigrationClient(cluster string) (*migrate.Migrate, error) {
	var dbType string
	var dbDriver database.Driver
	var cData clusterData
	switch cluster {
	case "mycluster":
		//cData = clusterMap[cluster]
		//pgConnectUrl = "postgres://%s:%s@%s/%s?sslmode=disable"
		//dbType = cData.dbType
		//dbUrl := fmt.Sprintf(pgConnectUrl, conf.dbUsername, conf.dbPassword,
		//	conf.dbHostPort, conf.dbName)
		//dbConn, err := sql.Open(dbType, dbUrl)
		//if err != nil {
		//	fmt.Println("Error connecting", err)
		//	return nil, err
		//}
		//dbDriver, err = postgres.WithInstance(dbConn, &postgres.Config{})
		//if err != nil {
		//	fmt.Println("Error initializing postgres driver ", err)
		//	return nil, err
		//}
	default:
		return nil, errors.New("unknown cluster")
	}

	return migrate.NewWithDatabaseInstance(cData.sourcePath, dbType, dbDriver)
}

func init() {

	clusterDataArray := []clusterData{
		{
			name:       "media",
			dbType:     "postgres",
			sourcePath: "file://migrator/files/",
			connConfig: connectConfig{ // Fetch config from config management, instead of hard-coded here
				dbUsername: "testUser",
				dbPassword: "testPass",
				dbHostPort: "192.186.100.10:5432",
				dbName:     "testDb",
			},
		},
	}
	clusterMap = make(map[string]clusterData, len(clusterDataArray))
	for _, cData := range clusterDataArray {
		clusterMap[cData.name] = cData
	}
}

func main() {
	migrator.Execute(getMigrationClient)
}
