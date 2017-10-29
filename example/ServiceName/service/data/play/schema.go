package main

import (
	"fmt"

	"github.com/carousell/DataAccessLayer/dal/cassandra"
	"github.com/carousell/Orion/example/ServiceName/service/data"
)

func main() {

	fmt.Println(cassandra.CreateCassandraTable(data.Message{}), ";")
}
