package miniorm

import (
	"context"
	"database/sql"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

//mockgen

//go:generate mockgen -source=$GOFILE -destination=$GOFILE.mock_test.go -package=$GOPACKAGE
type TableNameGetter interface {
	GetTableName() string
}

type IDGetter interface {
	GetID() (idColumn string, idValue int64)
}

type IDSetter interface {
	SetID(id int64)
}

type UniqueGetter interface {
	GetUniqueExpression() goqu.Ex
}

type OnCreator interface {
	OnCreate()
}

type OnUpdater interface {
	OnUpdate()
}

type QueryParams struct {
	TableName  string
	EntryList  interface{}
	Expression goqu.Expression
	OrderBy    []exp.OrderedExpression
	Limit      *uint32
	Offset     *uint32
}

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

type ORM interface {
	Create(ctx context.Context, entry interface{}) error
	Get(ctx context.Context, entry interface{}) error
	GetWithXLock(ctx context.Context, entry interface{}) error
	Query(ctx context.Context, params QueryParams) error
	QueryWithXLock(ctx context.Context, params QueryParams) error
	Count(ctx context.Context, tableName string, expression goqu.Expression) (int64, error)
	Update(ctx context.Context, entry interface{}) error
	CreateOrUpdate(ctx context.Context, entry interface{}) error
	Delete(ctx context.Context, entry interface{}) error
	GetDBWrapper() DBWrapper
	WithTx(executeFunc func(ORM) error) error
}

var (
	ErrNilEntry         = errors.New("entry is nil")
	ErrNotFound         = errors.New("entry not found")
	ErrUpdateNotApplied = errors.New("update not applied")
)

func NewORM(databaseConfig DatabaseConfig) (ORM, error) {
	switch databaseConfig.Driver {
	case DriverTypeMySQL:
		return NewMySQLORM(databaseConfig)
	case DriverTypePostgres:
		return NewPostgresORM(databaseConfig)
	case DriverTypeSQLite3:
		return NewSQLite3ORM(databaseConfig)
	case DriverTypeMSSQL:
		return NewMSSQLORM(databaseConfig)
	default:
		return nil, errors.New("invalid driver type")
	}
}
