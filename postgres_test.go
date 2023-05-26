package miniorm

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/assert"
)

var (
	postgresTestConfig = DatabaseConfig{
		Driver:       DriverTypePostgres,
		Host:         "localhost",
		Port:         5432,
		DatabaseName: "test",
		User:         "user",
		Password:     "password",
	}
)

func preparePostgresTestEntryTable(fixtureFile string) error {
	db, err := newSQLXDatabase(postgresTestConfig)
	if err != nil {
		return err
	}

	if _, err = db.Exec(`
		START TRANSACTION;

		DROP TABLE IF EXISTS get_id_entries;
		CREATE TABLE get_id_entries (
			id BIGSERIAL PRIMARY KEY,
			string_col TEXT NOT NULL,
			on_create_count BIGINT NOT NULL,
			on_update_count BIGINT NOT NULL
		);


		DROP TABLE IF EXISTS get_unique_entries;
		CREATE TABLE get_unique_entries (
			id_1 BIGINT,
			id_2 BIGINT,
			string_col TEXT NOT NULL,
			on_create_count BIGINT NOT NULL,
			on_update_count BIGINT NOT NULL,
			PRIMARY KEY (id_1, id_2)
		);

		END TRANSACTION;
	`); err != nil {
		return err
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db.DB),
		testfixtures.Dialect(string(DriverTypePostgres)),
		testfixtures.FilesMultiTables(fixtureFile),
	)
	if err != nil {
		return err
	}

	return fixtures.Load()
}

func TestPostgresCreate(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_create.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testCreate(t, orm, testfixturesDefaultSequenceStart)
}

func TestPostgresGet(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testGet(t, orm)
}

func TestPostgresGetWithXLock(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testGetWithXLock(t, orm)
}

func TestPostgresQuery(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testQuery(t, orm)
}

func TestPostgresQueryWithXLock(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testQueryWithXLock(t, orm)
}

func TestPostgresCount(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testCount(t, orm)
}

func TestPostgresCreateOrUpdate(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_create_or_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testCreateOrUpdate(t, orm, testfixturesDefaultSequenceStart)
}

func TestPostgresUpdate(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testUpdate(t, orm)
}

func TestPostgresDelete(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_delete.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testDelete(t, orm)
}

func TestPostgreSQLGetDBWrapper(t *testing.T) {
	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testGetDBWrapper(t, orm)
}

func TestPostgreSQLWithTX(t *testing.T) {
	err := preparePostgresTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(postgresTestConfig)
	assert.Nil(t, err)

	testWithTX(t, orm)
}
