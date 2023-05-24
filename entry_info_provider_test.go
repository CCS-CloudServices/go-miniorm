package miniorm

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestEntryInfoProviderGetEntryTableName(t *testing.T) {
	t.Parallel()

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	entryInfoProvider := newEntryInfoProvider()

	expectedTableName := "table name"
	tableNameGetter := NewMockTableNameGetter(mockController)
	tableNameGetter.EXPECT().GetTableName().Return(expectedTableName).Times(1)
	tableName, err := entryInfoProvider.GetEntryTableName(tableNameGetter)
	assert.Equal(t, expectedTableName, tableName)
	assert.Nil(t, err)

	testCaseList := []interface{}{
		1,
		1.0,
		"string",
		true,
		false,
		&struct{}{},
		struct{}{},
	}

	for _, testCase := range testCaseList {
		tableName, err := entryInfoProvider.GetEntryTableName(testCase)
		assert.Empty(t, tableName)
		assert.ErrorIs(t, ErrTableNameGetterExpected, err)
	}
}

func TestEntryInfoProviderGetID(t *testing.T) {
	t.Parallel()

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	entryInfoProvider := newEntryInfoProvider()

	expectedIDColumn := "column name 1"
	expectedIDValue := int64(1)
	idGetter := NewMockIDGetter(mockController)
	idGetter.EXPECT().GetID().Return(expectedIDColumn, expectedIDValue).Times(1)
	idColumn, idValue, err := entryInfoProvider.GetID(idGetter)
	assert.Equal(t, expectedIDColumn, idColumn)
	assert.NotZero(t, expectedIDValue, idValue)
	assert.Nil(t, err)

	testCaseList := []interface{}{
		1,
		1.0,
		"string",
		true,
		false,
		&struct{}{},
		struct{}{},
	}

	for _, testCase := range testCaseList {
		idColumn, idValue, err := entryInfoProvider.GetID(testCase)
		assert.Empty(t, idColumn)
		assert.Zero(t, idValue)
		assert.ErrorIs(t, ErrIDGetterExpected, err)
	}
}

func TestEntryInfoProviderGetEntrySelectExpression(t *testing.T) {
	t.Parallel()

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	entryInfoProvider := newEntryInfoProvider()

	expectedIDColumn := "column name 2"
	expectedIDValue := int64(1)
	expectedExpression := goqu.Ex{
		expectedIDColumn: expectedIDValue,
	}

	idGetter := NewMockIDGetter(mockController)
	idGetter.EXPECT().GetID().Return(expectedIDColumn, expectedIDValue).Times(1)
	expression, err := entryInfoProvider.GetEntrySelectExpression(idGetter)
	assert.Equal(t, expectedExpression, expression)
	assert.Nil(t, err)

	uniqueGetter := NewMockUniqueGetter(mockController)
	uniqueGetter.EXPECT().GetUniqueExpression().Return(expectedExpression).Times(1)
	expression, err = entryInfoProvider.GetEntrySelectExpression(uniqueGetter)
	assert.Equal(t, expectedExpression, expression)
	assert.Nil(t, err)

	testCaseList := []interface{}{
		1,
		1.0,
		"string",
		true,
		false,
		&struct{}{},
		struct{}{},
	}

	for _, testCase := range testCaseList {
		expression, err := entryInfoProvider.GetEntrySelectExpression(testCase)
		assert.Nil(t, expression)
		assert.ErrorIs(t, ErrUniqueGetterOrIDGetterExpected, err)
	}
}

func TestEntryInfoProviderOnCreateIfEntryIsOnCreator(t *testing.T) {
	t.Parallel()

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	entryInfoProvider := newEntryInfoProvider()

	onCreator := NewMockOnCreator(mockController)
	onCreator.EXPECT().OnCreate().Return().Times(1)
	assert.NotPanics(t, func() {
		entryInfoProvider.OnCreateIfEntryIsOnCreator(onCreator)
	})

	testCaseList := []interface{}{
		1,
		1.0,
		"string",
		true,
		false,
		&struct{}{},
		struct{}{},
	}

	for i := range testCaseList {
		testCase := testCaseList[i]
		assert.NotPanics(t, func() {
			entryInfoProvider.OnCreateIfEntryIsOnCreator(testCase)
		})
	}
}

func TestEntryInfoProviderOnCreateIfEntryIsOnUpdater(t *testing.T) {
	t.Parallel()

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	entryInfoProvider := newEntryInfoProvider()

	onUpdater := NewMockOnUpdater(mockController)
	onUpdater.EXPECT().OnUpdate().Return().Times(1)
	assert.NotPanics(t, func() {
		entryInfoProvider.OnUpdateIfEntryIsOnCreator(onUpdater)
	})

	testCaseList := []interface{}{
		1,
		1.0,
		"string",
		true,
		false,
		&struct{}{},
		struct{}{},
	}

	for i := range testCaseList {
		testCase := testCaseList[i]
		assert.NotPanics(t, func() {
			entryInfoProvider.OnUpdateIfEntryIsOnCreator(testCase)
		})
	}
}
