package miniorm

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/assert"
)

var (
	mssqlTestConfig = DatabaseConfig{
		Driver:       DriverTypeMSSQL,
		Host:         "localhost",
		Port:         1433,
		DatabaseName: "master",
		User:         "sa",
		Password:     "Acronis123",
	}
)

func prepareMSSQLTestEntryTable(fixtureFile string) error {
	db, err := newSQLXDatabase(mssqlTestConfig)
	if err != nil {
		return err
	}

	if _, err = db.Exec(`
		BEGIN TRANSACTION;

		IF OBJECT_ID('get_id_entries', 'U') IS NOT NULL
			DROP TABLE get_id_entries;
		CREATE TABLE get_id_entries (
			id BIGINT IDENTITY(1,1) PRIMARY KEY,
			string_col NVARCHAR(MAX) NOT NULL,
			on_create_count BIGINT NOT NULL,
			on_update_count BIGINT NOT NULL
		);

		IF OBJECT_ID('get_unique_entries', 'U') IS NOT NULL
			DROP TABLE get_unique_entries;
		CREATE TABLE get_unique_entries (
			id_1 BIGINT,
			id_2 BIGINT,
			string_col NVARCHAR(MAX) NOT NULL,
			on_create_count BIGINT NOT NULL,
			on_update_count BIGINT NOT NULL,
			CONSTRAINT PK_get_unique_entries PRIMARY KEY (id_1, id_2)
		);

		COMMIT TRANSACTION;
	`); err != nil {
		return err
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db.DB),
		testfixtures.Dialect(string(DriverTypeMSSQL)),
		testfixtures.FilesMultiTables(fixtureFile),
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
	if err != nil {
		return err
	}

	return fixtures.Load()
}

func TestMSSQLCreate(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_create.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testCreate(t, orm, 0)
}

func TestMSSQLGet(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testGet(t, orm)
}

func TestMSSQLGetWithXLock(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testGetWithXLock(t, orm)
}

func TestMssqlQuery(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testQuery(t, orm)
}

func TestMssqlQueryWithXLock(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testQueryWithXLock(t, orm)
}

func TestMssqlCount(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testCount(t, orm)
}

func TestMSSQLCreateOrUpdate(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_create_or_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	// MSSQL's auto increment ID takes the largest existing value + 1, so we set this value to be 100 to match the test data
	testCreateOrUpdate(t, orm, 100)
}

func TestMSSQLUpdate(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testUpdate(t, orm)
}

func TestMSSQLDelete(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_delete.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testDelete(t, orm)
}

func TestMSSQLGetDBWrapper(t *testing.T) {
	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testGetDBWrapper(t, orm)
}

func TestMSSQLWithTX(t *testing.T) {
	err := prepareMSSQLTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mssqlTestConfig)
	assert.Nil(t, err)

	testWithTX(t, orm)
}
