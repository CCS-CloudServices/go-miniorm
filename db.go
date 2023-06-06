package miniorm

import (
	"database/sql"
	"time"

	_ "github.com/denisenkom/go-mssqldb" // For MSSQL driver
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"     // For Mysql driver
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"  // For Postgres driver
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"   // For SQLite driver
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver" // For MSSQL driver
	_ "github.com/go-sql-driver/mysql"                   // For Mysql driver
	_ "github.com/jackc/pgx/v4/stdlib"                   // For Postgres driver
	_ "github.com/mattn/go-sqlite3"                      // For SQLite driver
)

var (
	configDriverTypeToDriverName = map[DriverType]string{
		DriverTypeMSSQL:    "sqlserver",
		DriverTypeMySQL:    "mysql",
		DriverTypePostgres: "pgx",
		DriverTypeSQLite3:  "sqlite3",
	}

	configDriverTypeToDialect = map[DriverType]string{
		DriverTypeMSSQL:    "sqlserver",
		DriverTypeMySQL:    "mysql",
		DriverTypePostgres: "postgres",
		DriverTypeSQLite3:  "sqlite3",
	}
)

func newSQLDatabase(databaseConfig DatabaseConfig) (*sql.DB, error) {
	sourceNameProvider, err := newSourceNameProvider(databaseConfig.Driver)
	if err != nil {
		return nil, err
	}

	sourceName, err := sourceNameProvider.GetSourceName(databaseConfig)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(configDriverTypeToDriverName[databaseConfig.Driver], sourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(databaseConfig.MaxOpenConnections)
	db.SetMaxIdleConns(databaseConfig.MaxIdleConnections)
	db.SetConnMaxIdleTime(time.Duration(databaseConfig.ConnMaxLifetimeInMinutes) * time.Minute)

	return db, nil
}

func newGoquDatabase(databaseConfig DatabaseConfig) (*goqu.Database, error) {
	db, err := newSQLDatabase(databaseConfig)
	if err != nil {
		return nil, err
	}

	return goqu.New(configDriverTypeToDialect[databaseConfig.Driver], db), nil
}
