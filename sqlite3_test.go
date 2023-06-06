package miniorm

import (
	"log"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/assert"
)

var (
	sqlite3TestConfigRetry = DatabaseConfig{
		Driver:                     DriverTypeSQLite3,
		URL:                        "file:test.db",
		SQLite3TransactionMode:     SQLite3TransactionModeRetry,
		SQLite3TransactionMaxRetry: 100,
		SQLite3TransactionRetryDelayInMillisecond:  100,
		SQLite3TransactionRetryJitterInMillisecond: 20,
		Logger: log.Default(),
	}

	sqlite3TestConfigMutex = DatabaseConfig{
		Driver:                 DriverTypeSQLite3,
		URL:                    "file:test.db",
		SQLite3TransactionMode: SQLite3TransactionModeMutex,
		Logger:                 log.Default(),
	}
)

func prepareSQLite3TestEntryTable(fixtureFile string) error {
	db, err := newSQLDatabase(sqlite3TestConfigMutex)
	if err != nil {
		return err
	}

	if _, err = db.Exec(`
		BEGIN TRANSACTION;

		DROP TABLE IF EXISTS get_id_entries;
		CREATE TABLE get_id_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			string_col TEXT NOT NULL,
			bytes_col BYTEA NOT NULL,
			on_create_count INTEGER NOT NULL,
			on_update_count INTEGER NOT NULL
		);

		DROP TABLE IF EXISTS get_unique_entries;
		CREATE TABLE get_unique_entries (
			id_1 INTEGER,
			id_2 INTEGER,
			string_col TEXT NOT NULL,
			bytes_col BYTEA NOT NULL,
			on_create_count INTEGER NOT NULL,
			on_update_count INTEGER NOT NULL,
			PRIMARY KEY (id_1, id_2)
		);

		COMMIT;
	`); err != nil {
		return err
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect(string(DriverTypeSQLite3)),
		testfixtures.FilesMultiTables(fixtureFile),
	)
	if err != nil {
		return err
	}

	return fixtures.Load()
}

func TestSQLite3CreateRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_create.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testCreate(t, orm, 0)
}

func TestSQLite3GetRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testGet(t, orm)
}

func TestSQLite3GetWithXLockRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testGetWithXLock(t, orm)
}

func TestSQLite3QueryRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testQuery(t, orm)
}

func TestSQLite3QueryWithXLockRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testQueryWithXLock(t, orm)
}

func TestSQLite3CountRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testCount(t, orm)
}

func TestSQLite3CreateOrUpdateRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_create_or_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	// SQLite's auto increment ID takes the largest existing value + 1, so we set this value to be 100 to match the test data
	testCreateOrUpdate(t, orm, 100)
}

func TestSQLite3UpdateRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testUpdate(t, orm)
}

func TestSQLite3DeleteRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_delete.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testDelete(t, orm)
}

func TestSQLite3GetDBWrapperRetry(t *testing.T) {
	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testGetDBWrapper(t, orm)
}

func TestSQLite3WithTXRetry(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigRetry)
	assert.Nil(t, err)

	testWithTX(t, orm)
}

func TestSQLite3CreateMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_create.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testCreate(t, orm, 0)
}

func TestSQLite3GetMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testGet(t, orm)
}

func TestSQLite3GetWithXLockMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testGetWithXLock(t, orm)
}

func TestSQLite3QueryMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testQuery(t, orm)
}

func TestSQLite3QueryWithXLockMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testQueryWithXLock(t, orm)
}

func TestSQLite3CountMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testCount(t, orm)
}

func TestSQLite3CreateOrUpdateMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_create_or_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	// SQLite's auto increment ID takes the largest existing value + 1, so we set this value to be 100 to match the test data
	testCreateOrUpdate(t, orm, 100)
}

func TestSQLite3UpdateMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testUpdate(t, orm)
}

func TestSQLite3DeleteMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_delete.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testDelete(t, orm)
}

func TestSQLite3GetDBWrapperMutex(t *testing.T) {
	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testGetDBWrapper(t, orm)
}

func TestSQLite3WithTXMutex(t *testing.T) {
	err := prepareSQLite3TestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(sqlite3TestConfigMutex)
	assert.Nil(t, err)

	testWithTX(t, orm)
}
