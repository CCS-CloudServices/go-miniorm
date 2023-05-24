package miniorm

import (
	"errors"
	"fmt"
)

type sourceNameProvider interface {
	GetSourceName(databaseConfig DatabaseConfig) (string, error)
}

type mssqlSourceNameProvider struct{}

func newMSSQLSourceNameProvider() sourceNameProvider {
	return &mssqlSourceNameProvider{}
}

func (*mssqlSourceNameProvider) GetSourceName(databaseConfig DatabaseConfig) (string, error) {
	if databaseConfig.Host == "" {
		return "", errors.New("host is not provided")
	}

	if databaseConfig.DatabaseName == "" {
		return "", errors.New("database name is not provided")
	}

	if databaseConfig.Port == 0 {
		return "", errors.New("port is not provided")
	}

	if databaseConfig.User == "" {
		return "", errors.New("user is not provided")
	}

	if databaseConfig.Password == "" {
		return "", errors.New("password is not provided")
	}

	return fmt.Sprintf(
		"sqlserver://%s:%s@%s:%v?database=%s",
		databaseConfig.User,
		databaseConfig.Password,
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.DatabaseName,
	), nil
}

type mysqlSourceNameProvider struct{}

func newMySQLSourceNameProvider() sourceNameProvider {
	return &mysqlSourceNameProvider{}
}

func (*mysqlSourceNameProvider) GetSourceName(databaseConfig DatabaseConfig) (string, error) {
	if databaseConfig.User == "" {
		return "", errors.New("user is not provided")
	}

	if databaseConfig.Password == "" {
		return "", errors.New("password is not provided")
	}

	if databaseConfig.Host == "" {
		return "", errors.New("host is not provided")
	}

	if databaseConfig.DatabaseName == "" {
		return "", errors.New("database name is not provided")
	}

	if databaseConfig.Port == 0 {
		return "", errors.New("port is not provided")
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		databaseConfig.User,
		databaseConfig.Password,
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.DatabaseName,
	), nil
}

type postgresSourceNameProvider struct{}

func newPostgresSourceNameProvider() sourceNameProvider {
	return &postgresSourceNameProvider{}
}

func (*postgresSourceNameProvider) GetSourceName(databaseConfig DatabaseConfig) (string, error) {
	if databaseConfig.Host == "" {
		return "", errors.New("host is not provided")
	}

	if databaseConfig.DatabaseName == "" {
		return "", errors.New("database name is not provided")
	}

	if databaseConfig.Port == 0 {
		return "", errors.New("port is not provided")
	}

	SourceName := fmt.Sprintf(
		"host=%s port=%v dbname=%s sslmode=disable",
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.DatabaseName,
	)

	if databaseConfig.User != "" {
		SourceName += fmt.Sprintf(" user=%s", databaseConfig.User)
	}

	if databaseConfig.Password != "" {
		SourceName += fmt.Sprintf(" password=%s", databaseConfig.Password)
	}

	return SourceName, nil
}

type sqlite3SourceNameProvider struct{}

func newSQLite3SourceNameProvider() sourceNameProvider {
	return &sqlite3SourceNameProvider{}
}

func (*sqlite3SourceNameProvider) GetSourceName(databaseConfig DatabaseConfig) (string, error) {
	if databaseConfig.URL == "" {
		return "", errors.New("url is not provided")
	}

	return databaseConfig.URL, nil
}

func newSourceNameProvider(driverType DriverType) (sourceNameProvider, error) {
	switch driverType {
	case DriverTypeMySQL:
		return newMySQLSourceNameProvider(), nil
	case DriverTypePostgres:
		return newPostgresSourceNameProvider(), nil
	case DriverTypeMSSQL:
		return newMSSQLSourceNameProvider(), nil
	case DriverTypeSQLite3:
		return newSQLite3SourceNameProvider(), nil
	default:
		return nil, errors.New("invalid driver type")
	}
}
