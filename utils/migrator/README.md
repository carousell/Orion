# Migrator

Migrator is an utility for schema management in microservices, which based on [github.com/golang-migrate/migrate](golang-migrate) and extended as an cli command using github.com/spf13/cobra

Given a cluster on which migration has to be managed - base engine has to be extended to enable execute of the supported commands.

## Usage
migration-client [flags] [command]
```
migration-client -c mysvcPG up -v 201903290057
```
###### Flag
-c or --cluster : cluster identifier on which migration has to be performed
-v or --version : version of the migration
###### Commands
- up : Applies all migration from current version of migration. If input version is specificed via version flag - applies migrations from current version until given version where given version should be greater than current.
- down : Applies migration from current version to input version specified via version flag. Here input version should lesser than current version
- force : Marks the version as applied without running any migration
- version : displays the current version of migration and its status - which signifies if the current version is applied successfully or not

## Extending

Please refer sample implementaion that extended `migratorcmdbase.go` in `example/migratorexample.go`

Note: vendor drivers from golang-migrate/migrate based on your requirement of dbType i.e postgres, cassandra, mysql etc,.