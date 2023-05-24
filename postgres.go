//nolint:dupl // Implementations of ORM for different database engines have many similarities
package miniorm

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type PostgresORM struct {
	db                DBWrapper
	entryInfoProvider *entryInfoProvider
}

func NewPostgresORM(databaseConfig DatabaseConfig) (ORM, error) {
	goquDB, err := newGoquDatabase(databaseConfig)
	if err != nil {
		return nil, err
	}

	return &PostgresORM{
		db:                goquDB,
		entryInfoProvider: newEntryInfoProvider(),
	}, nil
}

func (orm *PostgresORM) Create(ctx context.Context, entry interface{}) error {
	if entry == nil {
		return ErrNilEntry
	}

	orm.entryInfoProvider.OnCreateIfEntryIsOnCreator(entry)

	entryTableName, err := orm.entryInfoProvider.GetEntryTableName(entry)
	if err != nil {
		return err
	}

	insertDataset := orm.GetDBWrapper().Insert(entryTableName).Rows(entry)

	var (
		idColumn string
		idValue  int64
	)

	idSetterEntry, isIDSetterEntry := entry.(IDSetter)
	if isIDSetterEntry {
		idColumn, _, err = orm.entryInfoProvider.GetID(entry)
		if err != nil {
			return err
		}

		insertDataset = insertDataset.Returning(idColumn)
	}

	rows, err := insertDataset.Executor().QueryContext(ctx)
	if err != nil {
		return err
	}

	defer rows.Close()

	if isIDSetterEntry {
		if !rows.Next() {
			return ErrNotFound
		}

		if err := rows.Scan(&idValue); err != nil {
			return err
		}

		idSetterEntry.SetID(idValue)
	}

	return nil
}

func (orm *PostgresORM) CreateOrUpdate(ctx context.Context, entry interface{}) error {
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

		return txORM.Update(ctx, entry)
	})
}

func (orm *PostgresORM) Delete(ctx context.Context, entry interface{}) error {
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

func (orm *PostgresORM) Get(ctx context.Context, entry interface{}) error {
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

func (orm *PostgresORM) GetWithXLock(ctx context.Context, entry interface{}) error {
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

func (orm *PostgresORM) getQuerySelectDataset(params QueryParams) *goqu.SelectDataset {
	selectDataset := orm.db.Select().From(params.TableName).Where(params.Expression).Order(params.OrderBy...)

	if params.Offset != nil {
		selectDataset = selectDataset.Offset(uint(*params.Offset))
	}

	if params.Limit != nil {
		selectDataset = selectDataset.Limit(uint(*params.Limit))
	}

	return selectDataset
}

func (orm *PostgresORM) Query(ctx context.Context, params QueryParams) error {
	return orm.getQuerySelectDataset(params).ScanStructsContext(ctx, params.EntryList)
}

func (orm *PostgresORM) QueryWithXLock(ctx context.Context, params QueryParams) error {
	return orm.getQuerySelectDataset(params).ForUpdate(goqu.Wait).ScanStructsContext(ctx, params.EntryList)
}

func (orm *PostgresORM) Count(ctx context.Context, tableName string, expression exp.Expression) (int64, error) {
	count, err := orm.GetDBWrapper().Select().From(tableName).Where(expression).CountContext(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (orm *PostgresORM) Update(ctx context.Context, entry interface{}) error {
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

func (orm *PostgresORM) GetDBWrapper() DBWrapper {
	return orm.db
}

func (orm *PostgresORM) WithTx(executeFunc func(ORM) error) error {
	if nonTXDB, ok := orm.db.(*goqu.Database); ok {
		return nonTXDB.WithTx(func(td *goqu.TxDatabase) error {
			return executeFunc(&PostgresORM{
				db:                td,
				entryInfoProvider: orm.entryInfoProvider,
			})
		})
	}

	return executeFunc(orm)
}
