//nolint:dupl // Implementations of ORM for different database engines have many similarities
package miniorm

import (
	"context"
	"fmt"
	"regexp"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exec"
	"github.com/doug-martin/goqu/v9/exp"
)

type MSSQLORM struct {
	db                   DBWrapper
	entryInfoProvider    *entryInfoProvider
	insertIntoTableRegex *regexp.Regexp
	fromTableRegex       *regexp.Regexp
}

func NewMSSQLORM(databaseConfig DatabaseConfig) (ORM, error) {
	goquDB, err := newGoquDatabase(databaseConfig)
	if err != nil {
		return nil, err
	}

	goquDB.Logger(databaseConfig.Logger)

	return &MSSQLORM{
		db:                   goquDB,
		entryInfoProvider:    newEntryInfoProvider(),
		insertIntoTableRegex: regexp.MustCompile(`INSERT INTO "[^"]+"\s*\(("[^"]+",\s*)*("[^"]+")\)`),
		fromTableRegex:       regexp.MustCompile(`FROM\s+"[^"]+"`),
	}, nil
}

// HACK: Since goqu does not support MSSQL's OUTPUT syntax, we have to manually add that
func (orm *MSSQLORM) wrapInsertSQLStatementWithOutputIDColumn(statement, idColumn string) string {
	insertIntoTablePosition := orm.insertIntoTableRegex.FindStringIndex(statement)
	if insertIntoTablePosition == nil {
		return statement
	}

	return statement[:insertIntoTablePosition[1]] +
		fmt.Sprintf(" OUTPUT INSERTED.%s", idColumn) +
		statement[insertIntoTablePosition[1]:]
}

func (orm *MSSQLORM) Create(ctx context.Context, entry interface{}) error {
	if entry == nil {
		return ErrNilEntry
	}

	orm.entryInfoProvider.OnCreateIfEntryIsOnCreator(entry)

	entryTableName, err := orm.entryInfoProvider.GetEntryTableName(entry)
	if err != nil {
		return err
	}

	sqlStatement, params, err := orm.GetDBWrapper().
		Insert(entryTableName).
		Prepared(true).
		Rows(entry).
		ToSQL()
	if err != nil {
		return err
	}

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

		sqlStatement = orm.wrapInsertSQLStatementWithOutputIDColumn(sqlStatement, idColumn)
	}

	rows, err := orm.GetDBWrapper().QueryContext(ctx, sqlStatement, params...)
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

// HACK: Since goqu does not support MSSQL's lock syntax, we have to manually add that
func (orm *MSSQLORM) wrapSelectSQLStatementWithRowLock(statement string) string {
	fromTablePosition := orm.fromTableRegex.FindStringIndex(statement)
	if fromTablePosition == nil {
		return statement
	}

	return statement[:fromTablePosition[1]] + " WITH (XLOCK, ROWLOCK)" + statement[fromTablePosition[1]:]
}

func (orm *MSSQLORM) CreateOrUpdate(ctx context.Context, entry interface{}) error {
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

		sqlStatement, params, err := txORM.GetDBWrapper().
			Select().
			From(entryTableName).
			Where(selectEntryUniqueExpression).
			ToSQL()
		if err != nil {
			return err
		}

		rows, err := txORM.GetDBWrapper().QueryContext(ctx, orm.wrapSelectSQLStatementWithRowLock(sqlStatement), params...)
		if err != nil {
			return err
		}

		defer rows.Close()

		rowExists := rows.Next()
		if err := rows.Close(); err != nil {
			return err
		}

		if !rowExists {
			return txORM.Create(ctx, entry)
		}

		return txORM.Update(ctx, entry)
	})
}

func (orm *MSSQLORM) Delete(ctx context.Context, entry interface{}) error {
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
		Exec()
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

func (orm *MSSQLORM) Get(ctx context.Context, entry interface{}) error {
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

func (orm *MSSQLORM) GetWithXLock(ctx context.Context, entry interface{}) error {
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

	sqlQuery, params, err := orm.db.
		Select().
		From(entryTableName).
		Where(selectEntryUniqueExpression).
		Limit(1).
		ToSQL()
	if err != nil {
		return err
	}

	rows, err := orm.db.QueryContext(ctx, orm.wrapSelectSQLStatementWithRowLock(sqlQuery), params...)
	if err != nil {
		return err
	}

	defer rows.Close()

	if !rows.Next() {
		return ErrNotFound
	}

	return exec.NewScanner(rows).ScanStruct(entry)
}

func (orm *MSSQLORM) getQuerySelectDataset(params QueryParams) *goqu.SelectDataset {
	selectDataset := orm.db.Select().From(params.TableName).Where(params.Expression).Order(params.OrderBy...)

	if params.Offset != nil {
		selectDataset = selectDataset.Offset(uint(*params.Offset))
	}

	if params.Limit != nil {
		selectDataset = selectDataset.Limit(uint(*params.Limit))
	}

	return selectDataset
}

func (orm *MSSQLORM) Query(ctx context.Context, params QueryParams) error {
	return orm.getQuerySelectDataset(params).ScanStructsContext(ctx, params.EntryList)
}

func (orm *MSSQLORM) QueryWithXLock(ctx context.Context, params QueryParams) error {
	selectDataset := orm.getQuerySelectDataset(params)

	sqlStatement, sqlParams, err := selectDataset.ToSQL()
	if err != nil {
		return err
	}

	rows, err := orm.db.QueryContext(ctx, orm.wrapSelectSQLStatementWithRowLock(sqlStatement), sqlParams...)
	if err != nil {
		return err
	}

	defer rows.Close()

	return exec.NewScanner(rows).ScanStructs(params.EntryList)
}

func (orm *MSSQLORM) Count(ctx context.Context, tableName string, expression exp.Expression) (int64, error) {
	count, err := orm.db.Select().From(tableName).Where(expression).CountContext(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (orm *MSSQLORM) Update(ctx context.Context, entry interface{}) error {
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
		Prepared(true).
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

func (orm *MSSQLORM) GetDBWrapper() DBWrapper {
	return orm.db
}

func (orm *MSSQLORM) WithTx(executeFunc func(ORM) error) error {
	if nonTXDB, ok := orm.db.(*goqu.Database); ok {
		return nonTXDB.WithTx(func(td *goqu.TxDatabase) error {
			return executeFunc(&MSSQLORM{
				db:                   td,
				entryInfoProvider:    orm.entryInfoProvider,
				insertIntoTableRegex: orm.insertIntoTableRegex,
				fromTableRegex:       orm.fromTableRegex,
			})
		})
	}

	return executeFunc(orm)
}
