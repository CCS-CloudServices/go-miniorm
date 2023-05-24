package miniorm

type DriverType string
type SQLite3TransactionMode string

const (
	DriverTypeMySQL    DriverType = "mysql"
	DriverTypePostgres DriverType = "postgres"
	DriverTypeSQLite3  DriverType = "sqlite3"
	DriverTypeMSSQL    DriverType = "sqlserver"

	SQLite3TransactionModeRetry SQLite3TransactionMode = "retry"
	SQLite3TransactionModeMutex SQLite3TransactionMode = "mutex"
)

type DatabaseConfig struct {
	Driver                                     DriverType             `yaml:"driver"`
	Host                                       string                 `yaml:"host"`
	DatabaseName                               string                 `yaml:"databaseName"`
	Port                                       int                    `yaml:"port"`
	User                                       string                 `yaml:"user"`
	Password                                   string                 `yaml:"password"`
	URL                                        string                 `yaml:"url"`
	MaxOpenConnections                         int                    `yaml:"maxOpenConnections"`
	MaxIdleConnections                         int                    `yaml:"maxIdleConnections"`
	ConnMaxLifetimeInMinutes                   int                    `yaml:"connMaxLifetimeInMinutes"`
	SQLite3TransactionMode                     SQLite3TransactionMode `yaml:"sqlite3TransactionMode"`
	SQLite3TransactionMaxRetry                 uint                   `yaml:"sqlite3TransactionMaxRetry"`
	SQLite3TransactionRetryDelayInMillisecond  int                    `yaml:"sqlite3TransactionRetryDelayInMillisecond"`
	SQLite3TransactionRetryJitterInMillisecond int                    `yaml:"SQLite3TransactionRetryJitterInMillisecond"`
}
