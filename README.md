<a name="readme-top"></a>

<br />
<div align="center">
  <h3 align="center"><code>go-miniorm</code></h3>

  <p align="center">
    A simple wrapper over <a href="https://github.com/doug-martin/goqu">doug-martin/goqu</a> to simplify common SQL database operations
    <br />
  </p>
</div>

<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li>
      <a href="#usage">Usage</a>
          <ul>
              <li><a href="#defining-database-model">Defining database model</a></li>
              <li><a href="#initializing-the-orm">Initializing the ORM</a></li>
              <li><a href="#executing-database-operations">Executing database operations</a></li>
          </ul>
    </li>
    <li>
      <a href="#development">Development</a>
          <ul>
              <li><a href="#testing">Testing</a></li>
              <li><a href="#linting">Linting</a></li>
          </ul>
    </li>
  </ol>
</details>

## About The Project

Acronis uses 4 different SQL database engines in our services - MySQL, MSSQL, PostgreSQL, SQLite3. From our development experience, there are several cases where different logic are required for different database engine:

| Use case                              | MySQL's implementation with `goqu` | MSSQL's implementation with `goqu`                                                              | PostgreSQL's implementation with `goqu` | SQLite3's implementation with `goqu`                                                                                                            |
| ------------------------------------- | ---------------------------------- | ----------------------------------------------------------------------------------------------- | --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Retrieving IDs of newly inserted rows | Use `LastInsertId()`               | Make a second query for `SCOPE_IDENTITY()`                                                      | Use `RETURNING id`                      | Use `LastInsertId()`                                                                                                                            |
| Starting a transaction                | Use `WithTX()`                     | Use `WithTX()`                                                                                  | Use `WithTX()`                          | Use `WithTX()`, but [may yield errors due to `database is locked` error](https://github.com/mattn/go-sqlite3/issues/274#issuecomment-192131441) |
| Row locking in transactions           | Use `ForUpdate()`                  | Is not supported by `doug-martin/goqu`, need to use `SELECT ... FROM ... WITH (XLOCK, ROWLOCK)` | Use `ForUpdate()`                       | Is not supported by SQLite3                                                                                                                     |

`go-miniorm` provides a simple interface for common database operation, to simplify the process of working with different database engines.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Getting Started

### Prerequisites

Go 1.14 or above.

### Installation

```bash
go get git.acronis.com/libs/go-miniorm
```

## Usage

### Defining database model

The Go way of working with SQL database is with structs representing database model, using the tag `sql` to map from database columns to the appropriate struct fields:

```golang
type Entry struct {
	ID         int64  `db:"id"`
	StringCol  string `db:"string_col"`
	CreateTime int64  `db:"create_time"`
	UpdateTime int64  `db:"update_time"`
}
```

`go-miniorm` introduces 5 new interfaces, which can be implemented by model structs to be used in database operation:

#### `TableNameGetter`

`TableNameGetter` implements `GetTableName()`, which returns the table name containing the record:

```golang
type TableNameGetter interface {
	GetTableName() string
}

func (*Entry) GetTableName() string {
    return "entries"
}
```

If horizontal partitioning is used and records are distributed across multiple tables, you can also implement the logic to derive the table name here:

```golang
func (e *Entry) GetTableName() string {
    return fmt.Sprintf("entries_%d", e.ID % 1000)
}
```

Model structs are **required** to implement this interface.

#### `IDGetter`

`IDGetter` implements `GetID()`, which retrieves the integer ID of the record:

```golang
type IDGetter interface {
	GetID() (idColumn string, idValue int64)
}

func (e *Entry) GetID() (idColumn string, idValue int64) {
    return "id", e.ID
}
```

Model structs are **not required** to implement this interface. However, either `IDGetter` or `UniqueGetter` **must be implemented** for `go-miniorm` to execute retrieval/update operations (`Get()`, `Update()`, `Delete()`, `CreateOrUpdate()`) on existing records inside the database.

If both `IDGetter` and `UniqueGetter` are implemented, `IDGetter.GetID()` takes precedence.

#### `IDSetter`

`IDSetter` implements `SetID()`, which update the integer ID of the record:

```golang
type IDSetter interface {
	SetID(id int64)
}

func (e *Entry) SetID(id int64) {
    e.ID = id
}
```

Model structs are **not required** to implement this interface. However, if the database model uses `AUTO_INCREMENT` IDs (or equivalent), `IDSetter` **must be implemented** for `go-miniorm` to be able to retrieve and update the ID of newly created records (via `Create()` or `CreateOrUpdate()`).

#### `UniqueGetter`

`UniqueGetter` implements `GetUniqueExpression()`, which retrieves a `goqu.Ex` that can be used to uniquely identify the records inside the database:

```golang
type UniqueGetter interface {
	GetUniqueExpression() goqu.Ex
}

func (e *Entry) GetUniqueExpression() goqu.Ex {
    return goqu.Ex{
        "id": e.ID,
    }
}
```

This is useful when records are identified by multiple columns - for example, in N - N relationships or with `UNIQUE` constraints.

Model structs are **not required** to implement this interface. However, either `IDGetter` or `UniqueGetter` **must be implemented** for `go-miniorm` to execute retrieval/update operations (`Get()`, `Update()`, `Delete()`, `CreateOrUpdate()`) on existing records inside the database.

If both `IDGetter` and `UniqueGetter` are implemented, `IDGetter.GetID()` takes precedence.

#### `OnCreator`

`OnCreator` implements `OnCreate()`, which is a hook function that will be run before the record is created in the database (via `Create()` or `CreateOrUpdate()`, if the record didn't exist in the database beforehand):

```golang
type OnCreator interface {
	OnCreate()
}

func (e *Entry) OnCreate() {
    currentTime = time.Now().Unix()
    e.CreateTime = currentTime
    e.UpdateTime = currentTime
}
```

This is useful to execute certain updates before the record is inserted, with `Entry.CreateTime` above being an example.

Model structs are **not required** to implement this interface.

#### `OnUpdater`

`OnUpdater` implements `OnUpdate()`, which is a hook function that will be run before the record is updated in the database (via `Update()` or `CreateOrUpdate()`, if the record already existed in the database beforehand):

```golang
type OnUpdater interface {
	OnUpdate()
}

func (e *Entry) OnUpdate() {
    currentTime = time.Now().Unix()
    e.UpdateTime = currentTime
}
```

This is useful to execute certain updates before the record is updated, with `Entry.UpdateTime` above being an example.

Model structs are **not required** to implement this interface.

### Initializing the ORM

```golang
// For MySQL, Driver, Host, Port, DatabaseName, User and Password are required.
mysqlConfig := miniorm.DatabaseConfig{
    Driver:       miniorm.DriverTypeMySQL,
    Host:         "localhost",
    Port:         3306,
    DatabaseName: "test",
    User:         "root",
    Password:     "password",
}

// For MSSQL, Driver, Host, Port, DatabaseName, User and Password are required.
mssqlConfig = miniorm.DatabaseConfig{
    Driver:       miniorm.DriverTypeMSSQL,
    Host:         "localhost",
    Port:         1433,
    DatabaseName: "master",
    User:         "sa",
    Password:     "Acronis123",
}

// For PostgreSQL, Driver, Host, Port and DatabaseName are required. User and Password are optional.
postgresConfig = miniorm.DatabaseConfig{
    Driver:       miniorm.DriverTypePostgres,
    Host:         "localhost",
    Port:         5432,
    DatabaseName: "test",
    User:         "user",
    Password:     "password",
}

// For SQLite3, Driver and URL are required.
sqlite3ConfigRetry = miniorm.DatabaseConfig{
    Driver:                     miniorm.DriverTypeSQLite3,
    URL:                        "file:test.db",
    SQLite3TransactionMode:     miniorm.SQLite3TransactionModeRetry,
    SQLite3TransactionMaxRetry: 100,
    SQLite3TransactionRetryDelayInMillisecond:  100,
    SQLite3TransactionRetryJitterInMillisecond: 20,
}

sqlite3ConfigMutex = miniorm.DatabaseConfig{
    Driver:                 miniorm.DriverTypeSQLite3,
    URL:                    "file:test.db"
    SQLite3TransactionMode: miniorm.SQLite3TransactionModeMutex,
}

// Use miniorm.NewORM() if you want to derive the underlying database engine from databaseConfig
orm, err := miniorm.NewORM(mysqlConfig)

// Or use an explicit implementation
mySQLORM, err := miniorm.NewMySQLORM(mysqlConfig)
```

#### Configurations

| Field                                        | Type                                                 | Description                                                                                                                                          |
| -------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Driver`                                     | One of `mysql`, `sqlserver`, `postgres` or `sqlite3` | The database engine to connect to                                                                                                                    |
| `Host`                                       | string                                               | The host address of the database server (for MySQL, MSSQL and Postgres)                                                                              |
| `DatabaseName`                               | string                                               | The database name (for MySQL, MSSQL and Postgres)                                                                                                    |
| `Port`                                       | int                                                  | The port number of the database server (for MySQL, MSSQL and Postgres)                                                                               |
| `User`                                       | string                                               | The user on the database server (for MySQL, MSSQL and Postgres)                                                                                      |
| `Password`                                   | string                                               | The password of the user on the database server (for MySQL, MSSQL and Postgres)                                                                      |
| `URL`                                        | string                                               | The URL to the database file (for SQLite3)                                                                                                           |
| `MaxOpenConnections`                         | int                                                  | The maximum number of database connections to open in the connection pool                                                                            |
| `MaxIdleConnections`                         | int                                                  | The maximum number of idle database connections to be left in the connection pool                                                                    |
| `ConnMaxLifetimeInMinutes`                   | int                                                  | The maximum number of minute a database connection can stay idle before being closed                                                                 |
| `SQLite3TransactionMode`                     | One of `retry` or `mutex`                            | See <a href="#regarding-sqlite3transactionmode">Regarding `SQLite3TransactionMode`</a>                                                               |
| `SQLite3TransactionMaxRetry`                 | uint                                                 | If `SQLite3TransactionMode` is `retry`, the maximum number of retries when initiating a database transaction.                                        |
| `SQLite3TransactionRetryDelayInMillisecond`  | int                                                  | If `SQLite3TransactionMode` is `retry`, the delay (in milliseconds) between retries when initiating a database transaction.                          |
| `SQLite3TransactionRetryJitterInMillisecond` | int                                                  | If `SQLite3TransactionMode` is `retry`, the maximum random jitter in delay (in milliseconds) between retries when initiating a database transaction. |

#### Regarding `SQLite3TransactionMode`

Due to the nature of Golang's `sql.DB`, we cannot properly ensure that we only have one database connection to the SQLite's database file during transactions. Because of that, the `database is locked` error may occur when two or more transactions are requested at the same time, while the write lock to the database file is only provided for one.

Refer to [this GitHub issue of the Golang's SQLite3 driver](https://github.com/mattn/go-sqlite3/issues/274#issuecomment-192131441) for more details.

`go-miniorm` provides two approaches to mitigate this issue, configurable via the config `SQLite3TransactionMode`:

1. `SQLite3TransactionModeRetry`: In this mode, the same transaction is repeatedly retried until it is successful. There is a random delay in the range of `[delay - jitter, delay + jitter]` millisecond (both ends inclusive) between each attempt.
2. `SQLite3TransactionModeMutex`: In this mode, a global mutex is used to only allow one transaction from `go-miniorm` to be executed at all times.

`SQLite3TransactionModeRetry` allows Golang processes that use `go-miniorm` to share the same database file with those that don't, but incurs more performance penalty than `SQLite3TransactionModeMutex`. When writing a new service, prefer `SQLite3TransactionModeMutex` over `SQLite3TransactionModeRetry`, and make sure that different services use different database files, independent from each other.

### Executing database operations

#### `Create()`

```golang
entry := &Entry{
    StringCol: "value 1",
}
err := orm.Create(context.Background(), entry)
```

#### `Get()`

```golang
entry := &Entry{
    ID: 1,
}
err := orm.Get(context.Background(), entry)
```

#### `GetWithXLock()`

```golang
entry := &Entry{
    ID: 1,
}
txErr := orm.WithTx(func(o ORM) error {
    return o.GetWithXLock(context.Background(), entry)
})
```

#### `Query()`

```golang
err := orm.Query(context.Background(), miniorm.QueryParams{
    TableName:  "entry",
    Expression: goqu.Ex{},
    OrderBy: []exp.OrderedExpression{
        goqu.C("create_time").Desc(),
        goqu.C("id").Desc(),
    },
    Limit: proto.Uint32(100),
})
```

#### `QueryWithXLock()`

```golang
txErr := orm.WithTx(func (o ORM) error {
    return orm.QueryWithXLock(context.Background(), miniorm.QueryParams{
        TableName:  "entry",
        Expression: goqu.Ex{},
        OrderBy: []exp.OrderedExpression{
            goqu.C("create_time").Desc(),
            goqu.C("id").Desc(),
        },
        Limit: proto.Uint32(100),
    })
})
```

#### `Count()`

```golang
count, err := orm.Count(context.Background(), "entry", goqu.Ex{})
```

#### `Update()`

```golang
entry := &Entry{
    ID: 1,
    StringCol: "value 1",
}
err := orm.Update(context.Background(), entry)
```

#### `CreateOrUpdate()`

```golang
entry := &Entry{
    ID: 1,
    StringCol: "value 1",
}
err := orm.CreateOrUpdate(context.Background(), entry)
```

#### `Delete()`

```golang
entry := &Entry{
    ID: 1,
}
err := orm.Delete(context.Background(), entry)
```

#### `WithTx()`

Nested `WithTx()` calls are safe to use - if the ORM is already in a transaction, it will not start a new one.

```golang
txErr = orm.WithTx(func(o1 ORM) error {
    return o1.WithTx(func(o2 ORM) error {
        return o2.WithTx(func(o3 ORM) error {
            entry := &Entry{
                StringCol: "value 1",
            }
            return o3.Create(context.Background(), entry)
        })
    })
})
```

#### `GetDBWrapper()`

For more complex database operations such as `SELECT` with `JOIN`s, use `GetDBWrapper()` to get a simplified interface to interact with `goqu.DB` and `goqu.TxDB`:

```golang
type DBWrapper interface {
	From(cols ...interface{}) *goqu.SelectDataset
	Select(cols ...interface{}) *goqu.SelectDataset
	Update(table interface{}) *goqu.UpdateDataset
	Insert(table interface{}) *goqu.InsertDataset
	Delete(table interface{}) *goqu.DeleteDataset
	Truncate(table ...interface{}) *goqu.TruncateDataset
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

dbWrapper := orm.GetDBWrapper()
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Development

### Testing

```bash
# Start local database servers for testing purpose
make run-test-env

# In another terminal windows, execute the unit tests
make test
```

### Linting

```bash
make lint
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>
