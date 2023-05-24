package miniorm

import "github.com/doug-martin/goqu/v9"

const (
	getIDEntryTableName               = "get_id_entries"
	getIDEntryIDColumnName            = "id"
	getIDEntryOnCreateCountColumnName = "on_create_count"

	getUniqueEntryTableName     = "get_unique_entries"
	getUniqueEntryID1ColumnName = "id_1"
	getUniqueEntryID2ColumnName = "id_2"

	testfixturesDefaultSequenceStart = 10000
)

type getIDEntry struct {
	ID            int64  `db:"id" goqu:"skipinsert,skipupdate"`
	StringCol     string `db:"string_col"`
	OnCreateCount int64  `db:"on_create_count" goqu:"skipupdate"`
	OnUpdateCount int64  `db:"on_update_count"`
}

func (entry *getIDEntry) GetTableName() string {
	return getIDEntryTableName
}

func (entry *getIDEntry) GetID() (string, int64) {
	return getIDEntryIDColumnName, entry.ID
}

func (entry *getIDEntry) SetID(id int64) {
	entry.ID = id
}

type getIDEntryWithOnCreate struct {
	ID            int64  `db:"id" goqu:"skipinsert,skipupdate"`
	StringCol     string `db:"string_col"`
	OnCreateCount int64  `db:"on_create_count" goqu:"skipupdate"`
	OnUpdateCount int64  `db:"on_update_count"`
}

func (entry *getIDEntryWithOnCreate) GetTableName() string {
	return getIDEntryTableName
}

func (entry *getIDEntryWithOnCreate) GetID() (string, int64) {
	return getIDEntryIDColumnName, entry.ID
}

func (entry *getIDEntryWithOnCreate) SetID(id int64) {
	entry.ID = id
}

func (entry *getIDEntryWithOnCreate) OnCreate() {
	entry.OnCreateCount++
}

type getIDEntryWithOnUpdate struct {
	ID            int64  `db:"id" goqu:"skipinsert,skipupdate"`
	StringCol     string `db:"string_col"`
	OnCreateCount int64  `db:"on_create_count" goqu:"skipupdate"`
	OnUpdateCount int64  `db:"on_update_count"`
}

func (entry *getIDEntryWithOnUpdate) GetTableName() string {
	return getIDEntryTableName
}

func (entry *getIDEntryWithOnUpdate) GetID() (string, int64) {
	return getIDEntryIDColumnName, entry.ID
}

func (entry *getIDEntryWithOnUpdate) SetID(id int64) {
	entry.ID = id
}

func (entry *getIDEntryWithOnUpdate) OnUpdate() {
	entry.OnUpdateCount++
}

type getIDEntryWithOnCreateAndOnUpdate struct {
	ID            int64  `db:"id" goqu:"skipinsert,skipupdate"`
	StringCol     string `db:"string_col"`
	OnCreateCount int64  `db:"on_create_count" goqu:"skipupdate"`
	OnUpdateCount int64  `db:"on_update_count"`
}

func (entry *getIDEntryWithOnCreateAndOnUpdate) GetTableName() string {
	return getIDEntryTableName
}

func (entry *getIDEntryWithOnCreateAndOnUpdate) GetID() (string, int64) {
	return getIDEntryIDColumnName, entry.ID
}

func (entry *getIDEntryWithOnCreateAndOnUpdate) SetID(id int64) {
	entry.ID = id
}

func (entry *getIDEntryWithOnCreateAndOnUpdate) OnCreate() {
	entry.OnCreateCount++
}

func (entry *getIDEntryWithOnCreateAndOnUpdate) OnUpdate() {
	entry.OnUpdateCount++
}

type getUniqueEntryWithOnCreateAndOnUpdate struct {
	ID1           int64  `db:"id_1" goqu:"skipupdate"`
	ID2           int64  `db:"id_2" goqu:"skipupdate"`
	StringCol     string `db:"string_col"`
	OnCreateCount int64  `db:"on_create_count" goqu:"skipupdate"`
	OnUpdateCount int64  `db:"on_update_count"`
}

func (entry *getUniqueEntryWithOnCreateAndOnUpdate) GetTableName() string {
	return getUniqueEntryTableName
}

func (entry *getUniqueEntryWithOnCreateAndOnUpdate) GetUniqueExpression() goqu.Ex {
	return goqu.Ex{
		getUniqueEntryID1ColumnName: entry.ID1,
		getUniqueEntryID2ColumnName: entry.ID2,
	}
}

func (entry *getUniqueEntryWithOnCreateAndOnUpdate) OnCreate() {
	entry.OnCreateCount++
}

func (entry *getUniqueEntryWithOnCreateAndOnUpdate) OnUpdate() {
	entry.OnUpdateCount++
}

type tableNameGetterWithoutUniqueSelector struct{}

func (entry *tableNameGetterWithoutUniqueSelector) GetTableName() string {
	return ""
}
