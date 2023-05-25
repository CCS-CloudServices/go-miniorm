//nolint:dupl // Implementations of ORM for different database engines have many similarities
package miniorm

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

var (
	sqlite3ORMTxLock = new(sync.Mutex)
)

type SQLite3ORM struct {
	db                DBWrapper
	entryInfoProvider *entryInfoProvider
	databaseConfig    DatabaseConfig
}

func NewSQLite3ORM(databaseConfig DatabaseConfig) (ORM, error) {
	goquDB, err := newGoquDatabase(databaseConfig)
	if err != nil {
		return nil, err
	}

	return &SQLite3ORM{
		db:                goquDB,
		entryInfoProvider: newEntryInfoProvider(),
		databaseConfig:    databaseConfig,
	}, nil
}

func (orm *SQLite3ORM) Create(ctx context.Context, entry interface{}) error {
	if entry == nil {
		return ErrNilEntry
	}

	orm.entryInfoProvider.OnCreateIfEntryIsOnCreator(entry)

	entryTableName, err := orm.entryInfoProvider.GetEntryTableName(entry)
	if err != nil {
		return err
	}

	result, err := orm.GetDBWrapper().
		Insert(entryTableName).
		Rows(entry).
		Executor().
		ExecContext(ctx)
	if err != nil {
		return err
	}

	if idSetterEntry, ok := entry.(IDSetter); ok {
		entryID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		idSetterEntry.SetID(entryID)
	}

	return nil
}

func (orm *SQLite3ORM) CreateOrUpdate(ctx context.Context, entry interface{}) error {
	if entry == nil {
		return ErrNilEntry
	}

	return orm.WithTx(func(txORM ORM) error {
		entryTableName, err := orm.entryInfoProvider.GetEntryTableName(entry)
		if err != nil {
			return err
		}

		selectEntryUniqueExpression, err := orm.entryInfoProvider.GetEntrySelectExpression(entry)
		if err != nil {
			return err
		}

		count, err := txORM.GetDBWrapper().
			Select().
			From(entryTableName).
			Where(selectEntryUniqueExpression).
			CountContext(ctx)
		if err != nil {
			return err
		}

		if count == 0 {
			return txORM.Create(ctx, entry)
		}

		return txORM.Update(ctx, entry)
	})
}

func (orm *SQLite3ORM) Delete(ctx context.Context, entry interface{}) error {
	if entry == nil {
		return ErrNilEntry
	}

	entryTableName, err := orm.entryInfoProvider.GetEntryTableName(entry)
	if err != nil {
		return err
	}

	selectEntryUniqueExpression, err := orm.entryInfoProvider.GetEntrySelectExpression(entry)
	if err != nil {
		return err
	}

	result, err := orm.GetDBWrapper().
		Delete(entryTableName).
		Where(selectEntryUniqueExpression).
		Executor().
		ExecContext(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (orm *SQLite3ORM) Get(ctx context.Context, entry interface{}) error {
	if entry == nil {
		return ErrNilEntry
	}

	entryTableName, err := orm.entryInfoProvider.GetEntryTableName(entry)
	if err != nil {
		return err
	}

	selectEntryUniqueExpression, err := orm.entryInfoProvider.GetEntrySelectExpression(entry)
	if err != nil {
		return err
	}

	found, err := orm.GetDBWrapper().
		Select().
		From(entryTableName).
		Where(selectEntryUniqueExpression).
		Limit(1).
		ScanStructContext(ctx, entry)
	if err != nil {
		return err
	}

	if !found {
		return ErrNotFound
	}

	return nil
}

func (orm *SQLite3ORM) GetWithXLock(ctx context.Context, entry interface{}) error {
	// SQLite actually does not support row locking, so we just do a regular Get()
	return orm.Get(ctx, entry)
}

func (orm *SQLite3ORM) getQuerySelectDataset(params QueryParams) *goqu.SelectDataset {
	selectDataset := orm.GetDBWrapper().Select().From(params.TableName).Where(params.Expression).Order(params.OrderBy...)

	if params.Offset != nil {
		selectDataset = selectDataset.Offset(uint(*params.Offset))

		if params.Limit != nil {
			selectDataset = selectDataset.Limit(uint(*params.Limit))
		} else {
			selectDataset = selectDataset.Limit(math.MaxUint32)
		}
	} else if params.Limit != nil {
		selectDataset = selectDataset.Limit(uint(*params.Limit))
	}

	return selectDataset
}

func (orm *SQLite3ORM) Query(ctx context.Context, params QueryParams) error {
	return orm.getQuerySelectDataset(params).ScanStructsContext(ctx, params.EntryList)
}

func (orm *SQLite3ORM) QueryWithXLock(ctx context.Context, params QueryParams) error {
	// SQLite actually does not support row locking, so we just do a regular Query()
	return orm.Query(ctx, params)
}

func (orm *SQLite3ORM) Count(ctx context.Context, tableName string, expression exp.Expression) (int64, error) {
	count, err := orm.GetDBWrapper().Select().From(tableName).Where(expression).CountContext(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (orm *SQLite3ORM) Update(ctx context.Context, entry interface{}) error {
	if entry == nil {
		return ErrNilEntry
	}

	orm.entryInfoProvider.OnUpdateIfEntryIsOnCreator(entry)

	entryTableName, err := orm.entryInfoProvider.GetEntryTableName(entry)
	if err != nil {
		return err
	}

	selectEntryUniqueExpression, err := orm.entryInfoProvider.GetEntrySelectExpression(entry)
	if err != nil {
		return err
	}

	result, err := orm.GetDBWrapper().
		Update(entryTableName).
		Where(selectEntryUniqueExpression).
		Set(entry).
		Executor().
		ExecContext(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUpdateNotApplied
	}

	return nil
}

func (orm *SQLite3ORM) GetDBWrapper() DBWrapper {
	return orm.db
}

//nolint:gosec // Random sleep time does not need to be too secure
func (orm *SQLite3ORM) getRandomTxRetrySleepTime() time.Duration {
	sleepTimeInMillisecond := orm.databaseConfig.SQLite3TransactionRetryDelayInMillisecond -
		orm.databaseConfig.SQLite3TransactionRetryJitterInMillisecond +
		rand.Intn(orm.databaseConfig.SQLite3TransactionRetryJitterInMillisecond*2+1)

	return time.Duration(sleepTimeInMillisecond) * time.Millisecond
}

func (orm *SQLite3ORM) withTxRetry(nonTXDB *goqu.Database, executeFunc func(ORM) error) (err error) {
	for i := uint(0); i < orm.databaseConfig.SQLite3TransactionMaxRetry; i++ {
		err = nonTXDB.WithTx(func(td *goqu.TxDatabase) error {
			return executeFunc(&SQLite3ORM{
				db:                td,
				entryInfoProvider: orm.entryInfoProvider,
				databaseConfig:    orm.databaseConfig,
			})
		})

		if err == nil || errors.Is(err, ErrNilEntry) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrUpdateNotApplied) {
			return err
		}

		time.Sleep(orm.getRandomTxRetrySleepTime())
	}

	return err
}

// HACK: Due to the nature of Golang's sql.DB, we cannot properly ensure that we only have one database
// connection to the SQLite's database file during transaction. We use a client's side global mutex to
// prevent the "database is locked" error during transactions.
//
// Refer to https://github.com/mattn/go-sqlite3/issues/274#issuecomment-192131441.
func (orm *SQLite3ORM) withTxMutex(nonTXDB *goqu.Database, executeFunc func(ORM) error) error {
	sqlite3ORMTxLock.Lock()
	defer sqlite3ORMTxLock.Unlock()

	return nonTXDB.WithTx(func(td *goqu.TxDatabase) error {
		return executeFunc(&SQLite3ORM{
			db:                td,
			entryInfoProvider: orm.entryInfoProvider,
			databaseConfig:    orm.databaseConfig,
		})
	})
}

func (orm *SQLite3ORM) WithTx(executeFunc func(ORM) error) error {
	if nonTXDB, ok := orm.db.(*goqu.Database); ok {
		if orm.databaseConfig.SQLite3TransactionMode == SQLite3TransactionModeRetry {
			return orm.withTxRetry(nonTXDB, executeFunc)
		}

		return orm.withTxMutex(nonTXDB, executeFunc)
	}

	return executeFunc(orm)
}
