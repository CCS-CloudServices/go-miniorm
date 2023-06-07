package miniorm

import (
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql" // For Mysql driver
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/assert"
)

var (
	mysqlTestConfig = DatabaseConfig{
		Driver:       DriverTypeMySQL,
		Host:         "localhost",
		Port:         3306,
		DatabaseName: "test",
		User:         "root",
		Password:     "password",
		Logger:       log.Default(),
	}
)

func prepareMySQLTestEntryTable(fixtureFile string) error {
	db, err := newSQLDatabase(mysqlTestConfig)
	if err != nil {
		return err
	}

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS get_id_entries (
			id BIGINT NOT NULL AUTO_INCREMENT,
			string_col TEXT NOT NULL,
			bytes_col LONGBLOB NOT NULL,
			on_create_count BIGINT NOT NULL,
			on_update_count BIGINT NOT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB;
	`); err != nil {
		return err
	}

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS get_unique_entries (
			id_1 BIGINT,
			id_2 BIGINT,
			string_col TEXT NOT NULL,
			bytes_col LONGBLOB NOT NULL,
			on_create_count BIGINT NOT NULL,
			on_update_count BIGINT NOT NULL,
			PRIMARY KEY (id_1, id_2)
		) ENGINE=InnoDB;
	`); err != nil {
		return err
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect(string(DriverTypeMySQL)),
		testfixtures.FilesMultiTables(fixtureFile),
	)
	if err != nil {
		return err
	}

	return fixtures.Load()
}

func TestMySQLCreate(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_create.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	// MySQL's auto increment ID takes the current sequence value, so we had to decrease the starting value by one
	testCreate(t, orm, testfixturesDefaultSequenceStart-1)
}

func TestMySQLGet(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testGet(t, orm)
}

func TestMySQLGetWithXLock(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testGetWithXLock(t, orm)
}

func TestMySQLQuery(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testQuery(t, orm)
}

func TestMySQLQueryWithXLock(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testQueryWithXLock(t, orm)
}

func TestMySQLCount(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_query.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testCount(t, orm)
}

func TestMySQLCreateOrUpdate(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_create_or_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	// MySQL's auto increment ID takes the current sequence value, so we had to decrease the starting value by one
	testCreateOrUpdate(t, orm, testfixturesDefaultSequenceStart-1)
}

func TestMySQLUpdate(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_update.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testUpdate(t, orm)
}

func TestMySQLDelete(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_delete.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testDelete(t, orm)
}

func TestMySQLGetDBWrapper(t *testing.T) {
	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testGetDBWrapper(t, orm)
}

func TestMySQLWithTX(t *testing.T) {
	err := prepareMySQLTestEntryTable("testing/fixtures/test_get.yml")
	assert.Nil(t, err)

	orm, err := NewORM(mysqlTestConfig)
	assert.Nil(t, err)

	testWithTX(t, orm)
}
