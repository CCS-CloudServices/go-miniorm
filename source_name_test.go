package miniorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMSSQLProvider(t *testing.T) {
	t.Parallel()

	testCaseList := []struct {
		DatabaseConfig           DatabaseConfig
		ExpectedConnectionString string
		ExpectedNilErr           bool
	}{
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				Port:         1443,
				User:         "user",
				Password:     "password",
			},
			ExpectedConnectionString: "sqlserver://user:password@127.0.0.1:1443?database=database",
			ExpectedNilErr:           true,
		},
		{
			DatabaseConfig: DatabaseConfig{
				DatabaseName: "database",
				Port:         1443,
				User:         "user",
				Password:     "password",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:     "127.0.0.1",
				Port:     1443,
				User:     "user",
				Password: "password",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				User:         "user",
				Password:     "password",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				Port:         1443,
				Password:     "password",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				Port:         1443,
				User:         "user",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
	}

	for _, testCase := range testCaseList {
		mssqlProvider := newMSSQLSourceNameProvider()
		connectionString, err := mssqlProvider.GetSourceName(testCase.DatabaseConfig)
		assert.Equal(t, testCase.ExpectedConnectionString, connectionString)
		assert.Equal(t, testCase.ExpectedNilErr, err == nil)
	}
}

func TestMySQLProvider(t *testing.T) {
	t.Parallel()

	testCaseList := []struct {
		DatabaseConfig           DatabaseConfig
		ExpectedConnectionString string
		ExpectedNilErr           bool
	}{
		{
			DatabaseConfig:           DatabaseConfig{},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				User: "user",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				User:     "user",
				Password: "password",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				User:     "user",
				Password: "password",
				Host:     "127.0.0.1",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				User:         "user",
				Password:     "password",
				Host:         "127.0.0.1",
				DatabaseName: "test",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				User:         "user",
				Password:     "password",
				Host:         "127.0.0.1",
				DatabaseName: "test",
				Port:         3306,
			},
			ExpectedConnectionString: "user:password@tcp(127.0.0.1:3306)/test",
			ExpectedNilErr:           true,
		},
	}

	for _, testCase := range testCaseList {
		mysqlProvider := newMySQLSourceNameProvider()
		connectionString, err := mysqlProvider.GetSourceName(testCase.DatabaseConfig)
		assert.Equal(t, testCase.ExpectedConnectionString, connectionString)
		assert.Equal(t, testCase.ExpectedNilErr, err == nil)
	}
}

func TestPostgresProvider(t *testing.T) {
	t.Parallel()

	testCaseList := []struct {
		DatabaseConfig           DatabaseConfig
		ExpectedConnectionString string
		ExpectedNilErr           bool
	}{
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				Port:         1443,
			},
			ExpectedConnectionString: "host=127.0.0.1 port=1443 dbname=database sslmode=disable",
			ExpectedNilErr:           true,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				Port:         1443,
				User:         "user",
			},
			ExpectedConnectionString: "host=127.0.0.1 port=1443 dbname=database sslmode=disable user=user",
			ExpectedNilErr:           true,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				Port:         1443,
				Password:     "password",
			},
			ExpectedConnectionString: "host=127.0.0.1 port=1443 dbname=database sslmode=disable password=password",
			ExpectedNilErr:           true,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				Port:         1443,
				User:         "user",
				Password:     "password",
			},
			ExpectedConnectionString: "host=127.0.0.1 port=1443 dbname=database sslmode=disable user=user password=password",
			ExpectedNilErr:           true,
		},
		{
			DatabaseConfig: DatabaseConfig{
				DatabaseName: "database",
				Port:         1443,
				User:         "user",
				Password:     "password",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:     "127.0.0.1",
				Port:     1443,
				User:     "user",
				Password: "password",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
		{
			DatabaseConfig: DatabaseConfig{
				Host:         "127.0.0.1",
				DatabaseName: "database",
				User:         "user",
				Password:     "password",
			},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
	}

	for _, testCase := range testCaseList {
		postgresProvider := newPostgresSourceNameProvider()
		connectionString, err := postgresProvider.GetSourceName(testCase.DatabaseConfig)
		assert.Equal(t, testCase.ExpectedConnectionString, connectionString)
		assert.Equal(t, testCase.ExpectedNilErr, err == nil)
	}
}

func TestSQLite3Provider(t *testing.T) {
	t.Parallel()

	testCaseList := []struct {
		DatabaseConfig           DatabaseConfig
		ExpectedConnectionString string
		ExpectedNilErr           bool
	}{
		{
			DatabaseConfig: DatabaseConfig{
				URL: "./sqlite3.db",
			},
			ExpectedConnectionString: "./sqlite3.db",
			ExpectedNilErr:           true,
		},
		{
			DatabaseConfig:           DatabaseConfig{},
			ExpectedConnectionString: "",
			ExpectedNilErr:           false,
		},
	}

	for _, testCase := range testCaseList {
		sqliteProvider := newSQLite3SourceNameProvider()
		connectionString, err := sqliteProvider.GetSourceName(testCase.DatabaseConfig)
		assert.Equal(t, testCase.ExpectedConnectionString, connectionString)
		assert.Equal(t, testCase.ExpectedNilErr, err == nil)
	}
}
