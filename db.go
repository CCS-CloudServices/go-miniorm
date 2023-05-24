package miniorm

import (
	"time"

	_ "github.com/denisenkom/go-mssqldb" // For MSSQL driver
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"     // For Mysql driver
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"  // For Postgres driver
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"   // For SQLite driver
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver" // For MSSQL driver
	_ "github.com/go-sql-driver/mysql"                   // For Mysql driver
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // For Postgres driver
	_ "github.com/mattn/go-sqlite3" // For SQLite driver
)

func newSQLXDatabase(databaseConfig DatabaseConfig) (*sqlx.DB, error) {
	sourceNameProvider, err := newSourceNameProvider(databaseConfig.Driver)
	if err != nil {
		return nil, err
	}

	sourceName, err := sourceNameProvider.GetSourceName(databaseConfig)
	if err != nil {
		return nil, err
	}

	sqlxDatabase, err := sqlx.Connect(string(databaseConfig.Driver), sourceName)
	if err != nil {
		return nil, err
	}

	sqlxDatabase.SetMaxOpenConns(databaseConfig.MaxOpenConnections)
	sqlxDatabase.SetMaxIdleConns(databaseConfig.MaxIdleConnections)
	sqlxDatabase.SetConnMaxIdleTime(time.Duration(databaseConfig.ConnMaxLifetimeInMinutes) * time.Minute)

	return sqlxDatabase, nil
}

func newGoquDatabase(databaseConfig DatabaseConfig) (*goqu.Database, error) {
	sqlxDatabase, err := newSQLXDatabase(databaseConfig)
	if err != nil {
		return nil, err
	}

	return goqu.New(string(databaseConfig.Driver), sqlxDatabase), nil
}
