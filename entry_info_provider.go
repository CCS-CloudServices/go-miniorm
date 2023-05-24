package miniorm

import (
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

var (
	ErrTableNameGetterExpected        = errors.New("expected entry to implement TableNameGetter interface")
	ErrIDGetterExpected               = errors.New("expected entry to implement IDGetter interface")
	ErrUniqueGetterOrIDGetterExpected = errors.New("expected entry to implement UniqueGetter of IDGetter interface")
)

type entryInfoProvider struct{}

func newEntryInfoProvider() *entryInfoProvider {
	return &entryInfoProvider{}
}

func (*entryInfoProvider) GetEntryTableName(entry interface{}) (string, error) {
	tableNameGetterEntry, ok := entry.(TableNameGetter)
	if !ok {
		return "", ErrTableNameGetterExpected
	}

	return tableNameGetterEntry.GetTableName(), nil
}

func (*entryInfoProvider) GetID(entry interface{}) (idColumn string, idValue int64, err error) {
	idGetterEntry, ok := entry.(IDGetter)
	if !ok {
		return "", 0, ErrIDGetterExpected
	}

	columnName, id := idGetterEntry.GetID()

	return columnName, id, nil
}

func (manipulator *entryInfoProvider) GetEntrySelectExpression(entry interface{}) (exp.Ex, error) {
	uniqueGetter, ok := entry.(UniqueGetter)
	if ok {
		return uniqueGetter.GetUniqueExpression(), nil
	}

	idColumn, idValue, err := manipulator.GetID(entry)
	if err == nil {
		return goqu.Ex{idColumn: idValue}, nil
	}

	return nil, ErrUniqueGetterOrIDGetterExpected
}

func (*entryInfoProvider) OnCreateIfEntryIsOnCreator(entry interface{}) {
	onCreatorEntry, ok := entry.(OnCreator)
	if ok {
		onCreatorEntry.OnCreate()
	}
}

func (*entryInfoProvider) OnUpdateIfEntryIsOnCreator(entry interface{}) {
	onUpdaterEntry, ok := entry.(OnUpdater)
	if ok {
		onUpdaterEntry.OnUpdate()
	}
}
