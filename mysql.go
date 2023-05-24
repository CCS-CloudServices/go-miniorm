//nolint:dupl // Implementations of ORM for different database engines have many similarities
package miniorm

import (
	"context"
	"math"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type MySQLORM struct {
	db                DBWrapper
	entryInfoProvider *entryInfoProvider
}

func NewMySQLORM(databaseConfig DatabaseConfig) (ORM, error) {
	goquDB, err := newGoquDatabase(databaseConfig)
	if err != nil {
		return nil, err
	}

	return &MySQLORM{
		db:                goquDB,
		entryInfoProvider: newEntryInfoProvider(),
	}, nil
}

func (orm *MySQLORM) Create(ctx context.Context, entry interface{}) error {
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

func (orm *MySQLORM) CreateOrUpdate(ctx context.Context, entry interface{}) error {
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

		rows, err := txORM.GetDBWrapper().
			Select().
			From(entryTableName).
			Where(selectEntryUniqueExpression).
			ForUpdate(goqu.Wait).
			Executor().
			QueryContext(ctx)
		if err != nil {
			return err
		}

		defer rows.Close()

		if !rows.Next() {
			return txORM.Create(ctx, entry)
		}

		err = rows.Close()
		if err != nil {
			return err
		}

		return txORM.Update(ctx, entry)
	})
}

func (orm *MySQLORM) Delete(ctx context.Context, entry interface{}) error {
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

	result, err := orm.db.
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

func (orm *MySQLORM) Get(ctx context.Context, entry interface{}) error {
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

	found, err := orm.db.
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

func (orm *MySQLORM) GetWithXLock(ctx context.Context, entry interface{}) error {
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
		ForUpdate(goqu.Wait).
		Executor().
		ScanStructContext(ctx, entry)
	if err != nil {
		return err
	}

	if !found {
		return ErrNotFound
	}

	return nil
}

func (orm *MySQLORM) getQuerySelectDataset(params QueryParams) *goqu.SelectDataset {
	selectDataset := orm.db.Select().From(params.TableName).Where(params.Expression).Order(params.OrderBy...)

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

func (orm *MySQLORM) Query(ctx context.Context, params QueryParams) error {
	return orm.getQuerySelectDataset(params).ScanStructsContext(ctx, params.EntryList)
}

func (orm *MySQLORM) QueryWithXLock(ctx context.Context, params QueryParams) error {
	return orm.getQuerySelectDataset(params).ForUpdate(goqu.Wait).ScanStructsContext(ctx, params.EntryList)
}

func (orm *MySQLORM) Count(ctx context.Context, tableName string, expression exp.Expression) (int64, error) {
	count, err := orm.GetDBWrapper().Select().From(tableName).Where(expression).CountContext(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (orm *MySQLORM) Update(ctx context.Context, entry interface{}) error {
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

	result, err := orm.db.
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

func (orm *MySQLORM) GetDBWrapper() DBWrapper {
	return orm.db
}

func (orm *MySQLORM) WithTx(executeFunc func(ORM) error) error {
	if nonTXDB, ok := orm.db.(*goqu.Database); ok {
		return nonTXDB.WithTx(func(td *goqu.TxDatabase) error {
			return executeFunc(&MySQLORM{
				db:                td,
				entryInfoProvider: orm.entryInfoProvider,
			})
		})
	}

	return executeFunc(orm)
}
