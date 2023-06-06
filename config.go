package miniorm

type DriverType string
type SQLite3TransactionMode string

type Logger interface {
	Printf(format string, v ...interface{})
}

const (
	DriverTypeMySQL    DriverType = "mysql"
	DriverTypePostgres DriverType = "postgres"
	DriverTypeSQLite3  DriverType = "sqlite3"
	DriverTypeMSSQL    DriverType = "mssql"

	SQLite3TransactionModeRetry SQLite3TransactionMode = "retry"
	SQLite3TransactionModeMutex SQLite3TransactionMode = "mutex"
)

type DatabaseConfig struct {
	Driver                     DriverType             `yaml:"driver" json:"driver"`
	Host                       string                 `yaml:"host" json:"host"`
	DatabaseName               string                 `yaml:"databaseName" json:"databaseName"`
	Port                       int                    `yaml:"port" json:"port"`
	User                       string                 `yaml:"user" json:"user"`
	Password                   string                 `yaml:"password" json:"password"`
	URL                        string                 `yaml:"url" json:"url"`
	MaxOpenConnections         int                    `yaml:"maxOpenConnections" json:"maxOpenConnections"`
	MaxIdleConnections         int                    `yaml:"maxIdleConnections" json:"maxIdleConnections"`
	ConnMaxLifetimeInMinutes   int                    `yaml:"connMaxLifetimeInMinutes" json:"connMaxLifetimeInMinutes"`
	SQLite3TransactionMode     SQLite3TransactionMode `yaml:"sqlite3TransactionMode" json:"sqlite3TransactionMode"`
	SQLite3TransactionMaxRetry uint                   `yaml:"sqlite3TransactionMaxRetry" json:"sqlite3TransactionMaxRetry"`
	//nolint:lll // Long line, cannot be helped
	SQLite3TransactionRetryDelayInMillisecond int `yaml:"sqlite3TransactionRetryDelayInMillisecond" json:"sqlite3TransactionRetryDelayInMillisecond"`
	//nolint:lll // Long line, cannot be helped
	SQLite3TransactionRetryJitterInMillisecond int `yaml:"SQLite3TransactionRetryJitterInMillisecond" json:"SQLite3TransactionRetryJitterInMillisecond"`
	Logger                                     Logger
}
