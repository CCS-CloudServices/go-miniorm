package miniorm

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func testCreate(t *testing.T, orm ORM, sequenceStart int64) {
	err := orm.Create(context.Background(), nil)
	assert.ErrorIs(t, err, ErrNilEntry)

	err = orm.Create(context.Background(), struct{}{})
	assert.ErrorIs(t, err, ErrTableNameGetterExpected)

	getIDEntry1 := &getIDEntryWithOnCreateAndOnUpdate{
		StringCol: "value 1",
	}
	err = orm.Create(context.Background(), getIDEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            sequenceStart + 1,
		StringCol:     "value 1",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getIDEntry1)

	getIDEntry2 := &getIDEntryWithOnCreateAndOnUpdate{
		StringCol: "value 2",
	}
	err = orm.Create(context.Background(), getIDEntry2)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            sequenceStart + 2,
		StringCol:     "value 2",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getIDEntry2)

	getIDEntry3 := &getIDEntry{
		StringCol: "value 3",
	}
	err = orm.Create(context.Background(), getIDEntry3)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntry{
		ID:            sequenceStart + 3,
		StringCol:     "value 3",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}, getIDEntry3)

	getIDEntry4 := &getIDEntryWithOnCreate{
		StringCol: "value 4",
	}
	err = orm.Create(context.Background(), getIDEntry4)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreate{
		ID:            sequenceStart + 4,
		StringCol:     "value 4",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getIDEntry4)

	getIDEntry5 := &getIDEntryWithOnUpdate{
		StringCol: "value 5",
	}
	err = orm.Create(context.Background(), getIDEntry5)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnUpdate{
		ID:            sequenceStart + 5,
		StringCol:     "value 5",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}, getIDEntry5)

	getUniqueEntry1 := &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:       1,
		ID2:       2,
		StringCol: "value 1",
	}
	err = orm.Create(context.Background(), getUniqueEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:           1,
		ID2:           2,
		StringCol:     "value 1",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getUniqueEntry1)

	err = orm.Create(context.Background(), getUniqueEntry1)
	assert.NotNil(t, err)
}

func testGet(t *testing.T, orm ORM) {
	err := orm.Get(context.Background(), nil)
	assert.ErrorIs(t, err, ErrNilEntry)

	err = orm.Get(context.Background(), struct{}{})
	assert.ErrorIs(t, err, ErrTableNameGetterExpected)

	err = orm.Get(context.Background(), &tableNameGetterWithoutUniqueSelector{})
	assert.ErrorIs(t, err, ErrUniqueGetterOrIDGetterExpected)

	getIDEntry1 := &getIDEntryWithOnCreateAndOnUpdate{ID: 1}
	err = orm.Get(context.Background(), getIDEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            1,
		StringCol:     "value 1",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getIDEntry1)

	getIDEntry2 := &getIDEntryWithOnCreateAndOnUpdate{ID: 2}
	err = orm.Get(context.Background(), getIDEntry2)
	assert.ErrorIs(t, err, ErrNotFound)

	getUniqueEntry1 := &getUniqueEntryWithOnCreateAndOnUpdate{ID1: 1, ID2: 2}
	err = orm.Get(context.Background(), getUniqueEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:           1,
		ID2:           2,
		StringCol:     "value 1",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getUniqueEntry1)

	getUniqueEntry2 := &getUniqueEntryWithOnCreateAndOnUpdate{ID1: 2, ID2: 2}
	err = orm.Get(context.Background(), getUniqueEntry2)
	assert.ErrorIs(t, err, ErrNotFound)
}

func testGetWithXLock(t *testing.T, orm ORM) {
	err := orm.WithTx(func(o ORM) error {
		return orm.GetWithXLock(context.Background(), nil)
	})
	assert.ErrorIs(t, err, ErrNilEntry)

	err = orm.WithTx(func(o ORM) error {
		return orm.GetWithXLock(context.Background(), struct{}{})
	})
	assert.ErrorIs(t, err, ErrTableNameGetterExpected)

	err = orm.WithTx(func(o ORM) error {
		return orm.GetWithXLock(context.Background(), &tableNameGetterWithoutUniqueSelector{})
	})
	assert.ErrorIs(t, err, ErrUniqueGetterOrIDGetterExpected)

	updateCount := 64

	waitGroup1 := sync.WaitGroup{}
	for i := 0; i < updateCount; i++ {
		waitGroup1.Add(1)
		go func() {
			txErr := orm.WithTx(func(o ORM) error {
				getIDEntry1 := &getIDEntryWithOnCreateAndOnUpdate{ID: 1}
				if err := o.GetWithXLock(context.Background(), getIDEntry1); err != nil {
					return err
				}

				if err := o.Update(context.Background(), getIDEntry1); err != nil {
					return err
				}

				return nil
			})
			assert.Nil(t, txErr)
			waitGroup1.Done()
		}()
	}
	waitGroup1.Wait()

	finalGetIDEntry1 := &getIDEntryWithOnCreateAndOnUpdate{ID: 1}
	err = orm.Get(context.Background(), finalGetIDEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            1,
		StringCol:     "value 1",
		OnCreateCount: 1,
		OnUpdateCount: int64(updateCount),
	}, finalGetIDEntry1)

	waitGroup2 := sync.WaitGroup{}
	for i := 0; i < updateCount; i++ {
		waitGroup2.Add(1)
		go func() {
			txErr := orm.WithTx(func(o ORM) error {
				getIDEntry2 := &getIDEntryWithOnCreateAndOnUpdate{ID: 2}
				if err := o.GetWithXLock(context.Background(), getIDEntry2); err != nil {
					return err
				}

				return nil
			})
			assert.ErrorIs(t, txErr, ErrNotFound)
			waitGroup2.Done()
		}()
	}
	waitGroup2.Wait()

	waitGroup3 := sync.WaitGroup{}
	for i := 0; i < updateCount; i++ {
		waitGroup3.Add(1)
		go func() {
			txErr := orm.WithTx(func(o ORM) error {
				getUniqueEntry1 := &getUniqueEntryWithOnCreateAndOnUpdate{ID1: 1, ID2: 2}
				if err := o.GetWithXLock(context.Background(), getUniqueEntry1); err != nil {
					return err
				}

				if err := o.Update(context.Background(), getUniqueEntry1); err != nil {
					return err
				}

				return nil
			})
			assert.Nil(t, txErr)
			waitGroup3.Done()
		}()
	}
	waitGroup3.Wait()

	finalGetUniqueEntry1 := &getUniqueEntryWithOnCreateAndOnUpdate{ID1: 1, ID2: 2}
	err = orm.Get(context.Background(), finalGetUniqueEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:           1,
		ID2:           2,
		StringCol:     "value 1",
		OnCreateCount: 1,
		OnUpdateCount: int64(updateCount),
	}, finalGetUniqueEntry1)

	waitGroup4 := sync.WaitGroup{}
	for i := 0; i < updateCount; i++ {
		waitGroup4.Add(1)
		go func() {
			txErr := orm.WithTx(func(o ORM) error {
				getUniqueEntry2 := &getUniqueEntryWithOnCreateAndOnUpdate{ID1: 2, ID2: 2}
				if err := o.GetWithXLock(context.Background(), getUniqueEntry2); err != nil {
					return err
				}

				return nil
			})
			assert.ErrorIs(t, txErr, ErrNotFound)
			waitGroup4.Done()
		}()
	}
	waitGroup4.Wait()
}

func testQuery(t *testing.T, orm ORM) {
	testCases := []struct {
		Params            QueryParams
		ExpectedEntryList []getIDEntryWithOnCreateAndOnUpdate
		MatchOrder        bool
	}{
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 1, StringCol: "value 1", OnCreateCount: 1, OnUpdateCount: 0},
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: 0},
				{ID: 3, StringCol: "value 3", OnCreateCount: 1, OnUpdateCount: 0},
				{ID: 4, StringCol: "value 4", OnCreateCount: 10, OnUpdateCount: 0},
				{ID: 5, StringCol: "value 5", OnCreateCount: 10, OnUpdateCount: 0},
			},
			MatchOrder: false,
		},
		{
			Params: QueryParams{
				TableName: getIDEntryTableName,
				Expression: goqu.And(
					goqu.C(getIDEntryOnCreateCountColumnName).Lt(10),
				),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 1, StringCol: "value 1", OnCreateCount: 1, OnUpdateCount: 0},
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: 0},
				{ID: 3, StringCol: "value 3", OnCreateCount: 1, OnUpdateCount: 0},
			},
			MatchOrder: false,
		},
		{
			Params: QueryParams{
				TableName: getIDEntryTableName,
				Expression: goqu.And(
					goqu.C(getIDEntryOnCreateCountColumnName).Gte(10),
				),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 4, StringCol: "value 4", OnCreateCount: 10, OnUpdateCount: 0},
				{ID: 5, StringCol: "value 5", OnCreateCount: 10, OnUpdateCount: 0},
			},
			MatchOrder: false,
		},
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
				OrderBy: []exp.OrderedExpression{
					goqu.C(getIDEntryOnCreateCountColumnName).Desc(),
					goqu.C(getIDEntryIDColumnName).Desc(),
				},
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 5, StringCol: "value 5", OnCreateCount: 10, OnUpdateCount: 0},
				{ID: 4, StringCol: "value 4", OnCreateCount: 10, OnUpdateCount: 0},
				{ID: 3, StringCol: "value 3", OnCreateCount: 1, OnUpdateCount: 0},
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: 0},
				{ID: 1, StringCol: "value 1", OnCreateCount: 1, OnUpdateCount: 0},
			},
			MatchOrder: true,
		},
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
				OrderBy: []exp.OrderedExpression{
					goqu.C(getIDEntryOnCreateCountColumnName).Desc(),
					goqu.C(getIDEntryIDColumnName).Desc(),
				},
				Offset: proto.Uint32(3),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: 0},
				{ID: 1, StringCol: "value 1", OnCreateCount: 1, OnUpdateCount: 0},
			},
			MatchOrder: true,
		},
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
				OrderBy: []exp.OrderedExpression{
					goqu.C(getIDEntryOnCreateCountColumnName).Desc(),
					goqu.C(getIDEntryIDColumnName).Desc(),
				},
				Limit: proto.Uint32(3),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 5, StringCol: "value 5", OnCreateCount: 10, OnUpdateCount: 0},
				{ID: 4, StringCol: "value 4", OnCreateCount: 10, OnUpdateCount: 0},
				{ID: 3, StringCol: "value 3", OnCreateCount: 1, OnUpdateCount: 0},
			},
			MatchOrder: true,
		},
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
				OrderBy: []exp.OrderedExpression{
					goqu.C(getIDEntryOnCreateCountColumnName).Desc(),
					goqu.C(getIDEntryIDColumnName).Desc(),
				},
				Offset: proto.Uint32(3),
				Limit:  proto.Uint32(1),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: 0},
			},
			MatchOrder: true,
		},
	}

	for _, testCase := range testCases {
		params := testCase.Params
		entryList := make([]getIDEntryWithOnCreateAndOnUpdate, 0)
		params.EntryList = &entryList
		err := orm.Query(context.Background(), params)
		assert.Nil(t, err)

		if testCase.MatchOrder {
			assert.Equal(t, testCase.ExpectedEntryList, entryList)
		} else {
			assert.ElementsMatch(t, testCase.ExpectedEntryList, entryList)
		}
	}
}

func testQueryWithXLock(t *testing.T, orm ORM) {
	updateCount := int64(64)

	testCases := []struct {
		Params            QueryParams
		ExpectedEntryList []getIDEntryWithOnCreateAndOnUpdate
		MatchOrder        bool
	}{
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 1, StringCol: "value 1", OnCreateCount: 1, OnUpdateCount: updateCount},
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: updateCount},
				{ID: 3, StringCol: "value 3", OnCreateCount: 1, OnUpdateCount: updateCount},
				{ID: 4, StringCol: "value 4", OnCreateCount: 10, OnUpdateCount: updateCount},
				{ID: 5, StringCol: "value 5", OnCreateCount: 10, OnUpdateCount: updateCount},
			},
			MatchOrder: false,
		},
		{
			Params: QueryParams{
				TableName: getIDEntryTableName,
				Expression: goqu.And(
					goqu.C(getIDEntryOnCreateCountColumnName).Lt(10),
				),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 1, StringCol: "value 1", OnCreateCount: 1, OnUpdateCount: updateCount * 2},
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: updateCount * 2},
				{ID: 3, StringCol: "value 3", OnCreateCount: 1, OnUpdateCount: updateCount * 2},
			},
			MatchOrder: false,
		},
		{
			Params: QueryParams{
				TableName: getIDEntryTableName,
				Expression: goqu.And(
					goqu.C(getIDEntryOnCreateCountColumnName).Gte(10),
				),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 4, StringCol: "value 4", OnCreateCount: 10, OnUpdateCount: updateCount * 2},
				{ID: 5, StringCol: "value 5", OnCreateCount: 10, OnUpdateCount: updateCount * 2},
			},
			MatchOrder: false,
		},
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
				OrderBy: []exp.OrderedExpression{
					goqu.C(getIDEntryOnCreateCountColumnName).Desc(),
					goqu.C(getIDEntryIDColumnName).Desc(),
				},
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 5, StringCol: "value 5", OnCreateCount: 10, OnUpdateCount: updateCount * 3},
				{ID: 4, StringCol: "value 4", OnCreateCount: 10, OnUpdateCount: updateCount * 3},
				{ID: 3, StringCol: "value 3", OnCreateCount: 1, OnUpdateCount: updateCount * 3},
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: updateCount * 3},
				{ID: 1, StringCol: "value 1", OnCreateCount: 1, OnUpdateCount: updateCount * 3},
			},
			MatchOrder: true,
		},
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
				OrderBy: []exp.OrderedExpression{
					goqu.C(getIDEntryOnCreateCountColumnName).Desc(),
					goqu.C(getIDEntryIDColumnName).Desc(),
				},
				Offset: proto.Uint32(3),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: updateCount * 4},
				{ID: 1, StringCol: "value 1", OnCreateCount: 1, OnUpdateCount: updateCount * 4},
			},
			MatchOrder: true,
		},
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
				OrderBy: []exp.OrderedExpression{
					goqu.C(getIDEntryOnCreateCountColumnName).Desc(),
					goqu.C(getIDEntryIDColumnName).Desc(),
				},
				Limit: proto.Uint32(3),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 5, StringCol: "value 5", OnCreateCount: 10, OnUpdateCount: updateCount * 4},
				{ID: 4, StringCol: "value 4", OnCreateCount: 10, OnUpdateCount: updateCount * 4},
				{ID: 3, StringCol: "value 3", OnCreateCount: 1, OnUpdateCount: updateCount * 4},
			},
			MatchOrder: true,
		},
		{
			Params: QueryParams{
				TableName:  getIDEntryTableName,
				Expression: goqu.Ex{},
				OrderBy: []exp.OrderedExpression{
					goqu.C(getIDEntryOnCreateCountColumnName).Desc(),
					goqu.C(getIDEntryIDColumnName).Desc(),
				},
				Offset: proto.Uint32(3),
				Limit:  proto.Uint32(1),
			},
			ExpectedEntryList: []getIDEntryWithOnCreateAndOnUpdate{
				{ID: 2, StringCol: "value 2", OnCreateCount: 1, OnUpdateCount: updateCount * 5},
			},
			MatchOrder: true,
		},
	}

	for _, testCase := range testCases {
		waitGroup := sync.WaitGroup{}
		for i := 0; i < int(updateCount); i++ {
			params := testCase.Params
			entryList := make([]getIDEntryWithOnCreateAndOnUpdate, 0)
			params.EntryList = &entryList

			waitGroup.Add(1)
			go func() {
				txErr := orm.WithTx(func(o ORM) error {
					if err := o.QueryWithXLock(context.Background(), params); err != nil {
						return err
					}

					for _, entry := range entryList {
						if err := o.Update(context.Background(), &entry); err != nil {
							return err
						}
					}

					return nil
				})
				assert.Nil(t, txErr)
				waitGroup.Done()
			}()
		}
		waitGroup.Wait()

		params := testCase.Params
		entryList := make([]getIDEntryWithOnCreateAndOnUpdate, 0)
		params.EntryList = &entryList
		err := orm.Query(context.Background(), params)
		assert.Nil(t, err)

		if testCase.MatchOrder {
			assert.Equal(t, testCase.ExpectedEntryList, entryList)
		} else {
			assert.ElementsMatch(t, testCase.ExpectedEntryList, entryList)
		}
	}
}

func testCount(t *testing.T, orm ORM) {
	testCases := []struct {
		Expression    goqu.Expression
		ExpectedCount int64
	}{
		{
			Expression:    goqu.Ex{},
			ExpectedCount: 5,
		},
		{
			Expression: goqu.And(
				goqu.C(getIDEntryOnCreateCountColumnName).Lt(10),
			),
			ExpectedCount: 3,
		},
		{
			Expression: goqu.And(
				goqu.C(getIDEntryOnCreateCountColumnName).Gte(10),
			),
			ExpectedCount: 2,
		},
	}

	for _, testCase := range testCases {
		count, err := orm.Count(context.Background(), getIDEntryTableName, testCase.Expression)
		assert.Equal(t, testCase.ExpectedCount, count)
		assert.Nil(t, err)
	}
}

func testCreateOrUpdate(t *testing.T, orm ORM, sequenceStart int64) {
	err := orm.CreateOrUpdate(context.Background(), nil)
	assert.ErrorIs(t, err, ErrNilEntry)

	err = orm.CreateOrUpdate(context.Background(), struct{}{})
	assert.ErrorIs(t, err, ErrTableNameGetterExpected)

	err = orm.CreateOrUpdate(context.Background(), &tableNameGetterWithoutUniqueSelector{})
	assert.ErrorIs(t, err, ErrUniqueGetterOrIDGetterExpected)

	getIDEntry1 := &getIDEntryWithOnCreateAndOnUpdate{
		ID:            100,
		StringCol:     "value 1",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.CreateOrUpdate(context.Background(), getIDEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            100,
		StringCol:     "value 1",
		OnCreateCount: 0,
		OnUpdateCount: 1,
	}, getIDEntry1)

	getIDEntry2 := &getIDEntry{
		ID:            100,
		StringCol:     "value 2",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.CreateOrUpdate(context.Background(), getIDEntry2)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntry{
		ID:            100,
		StringCol:     "value 2",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}, getIDEntry2)

	getIDEntry3 := &getIDEntryWithOnCreate{
		ID:            100,
		StringCol:     "value 3",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.CreateOrUpdate(context.Background(), getIDEntry3)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreate{
		ID:            100,
		StringCol:     "value 3",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}, getIDEntry3)

	getIDEntry4 := &getIDEntryWithOnUpdate{
		ID:            100,
		StringCol:     "value 4",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.CreateOrUpdate(context.Background(), getIDEntry4)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnUpdate{
		ID:            100,
		StringCol:     "value 4",
		OnCreateCount: 0,
		OnUpdateCount: 1,
	}, getIDEntry4)

	getIDEntry5 := &getIDEntryWithOnCreateAndOnUpdate{
		ID:            2,
		StringCol:     "value 2",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.CreateOrUpdate(context.Background(), getIDEntry5)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            sequenceStart + 1,
		StringCol:     "value 2",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getIDEntry5)

	getIDEntry6 := &getIDEntry{
		ID:            3,
		StringCol:     "value 3",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.CreateOrUpdate(context.Background(), getIDEntry6)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntry{
		ID:            sequenceStart + 2,
		StringCol:     "value 3",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}, getIDEntry6)

	getIDEntry7 := &getIDEntryWithOnCreate{
		ID:            4,
		StringCol:     "value 4",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.CreateOrUpdate(context.Background(), getIDEntry7)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreate{
		ID:            sequenceStart + 3,
		StringCol:     "value 4",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getIDEntry7)

	getIDEntry8 := &getIDEntryWithOnUpdate{
		ID:        5,
		StringCol: "value 5",
	}
	err = orm.CreateOrUpdate(context.Background(), getIDEntry8)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnUpdate{
		ID:            sequenceStart + 4,
		StringCol:     "value 5",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}, getIDEntry8)

	getUniqueEntry1 := &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:       1,
		ID2:       2,
		StringCol: "value 1",
	}
	err = orm.CreateOrUpdate(context.Background(), getUniqueEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:           1,
		ID2:           2,
		StringCol:     "value 1",
		OnCreateCount: 0,
		OnUpdateCount: 1,
	}, getUniqueEntry1)

	getUniqueEntry2 := &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:       2,
		ID2:       2,
		StringCol: "value 2",
	}
	err = orm.CreateOrUpdate(context.Background(), getUniqueEntry2)
	assert.Nil(t, err)
	assert.Equal(t, &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:           2,
		ID2:           2,
		StringCol:     "value 2",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, getUniqueEntry2)
}

func testUpdate(t *testing.T, orm ORM) {
	err := orm.Update(context.Background(), nil)
	assert.ErrorIs(t, err, ErrNilEntry)

	err = orm.Update(context.Background(), struct{}{})
	assert.ErrorIs(t, err, ErrTableNameGetterExpected)

	err = orm.Update(context.Background(), &tableNameGetterWithoutUniqueSelector{})
	assert.ErrorIs(t, err, ErrUniqueGetterOrIDGetterExpected)

	getIDEntry1 := &getIDEntryWithOnCreateAndOnUpdate{
		ID:            1,
		StringCol:     "value 1",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.Update(context.Background(), getIDEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            1,
		StringCol:     "value 1",
		OnCreateCount: 0,
		OnUpdateCount: 1,
	}, getIDEntry1)

	getIDEntry2 := &getIDEntry{
		ID:            1,
		StringCol:     "value 2",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.Update(context.Background(), getIDEntry2)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntry{
		ID:            1,
		StringCol:     "value 2",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}, getIDEntry2)

	getIDEntry3 := &getIDEntryWithOnCreate{
		ID:            1,
		StringCol:     "value 3",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.Update(context.Background(), getIDEntry3)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreate{
		ID:            1,
		StringCol:     "value 3",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}, getIDEntry3)

	getIDEntry4 := &getIDEntryWithOnUpdate{
		ID:            1,
		StringCol:     "value 4",
		OnCreateCount: 0,
		OnUpdateCount: 0,
	}
	err = orm.Update(context.Background(), getIDEntry4)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnUpdate{
		ID:            1,
		StringCol:     "value 4",
		OnCreateCount: 0,
		OnUpdateCount: 1,
	}, getIDEntry4)

	getIDEntry5 := &getIDEntryWithOnCreateAndOnUpdate{
		ID:        2,
		StringCol: "value 2",
	}
	err = orm.Update(context.Background(), getIDEntry5)
	assert.ErrorIs(t, err, ErrUpdateNotApplied)

	getUniqueEntry1 := &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:       1,
		ID2:       2,
		StringCol: "value 1",
	}
	err = orm.Update(context.Background(), getUniqueEntry1)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            1,
		StringCol:     "value 1",
		OnCreateCount: 0,
		OnUpdateCount: 1,
	}, getIDEntry1)

	getUniqueEntry2 := &getUniqueEntryWithOnCreateAndOnUpdate{
		ID1:       2,
		ID2:       2,
		StringCol: "value 2",
	}
	err = orm.Update(context.Background(), getUniqueEntry2)
	assert.ErrorIs(t, err, ErrUpdateNotApplied)
}

func testDelete(t *testing.T, orm ORM) {
	err := orm.Delete(context.Background(), nil)
	assert.ErrorIs(t, err, ErrNilEntry)

	err = orm.Delete(context.Background(), struct{}{})
	assert.ErrorIs(t, err, ErrTableNameGetterExpected)

	err = orm.Delete(context.Background(), &tableNameGetterWithoutUniqueSelector{})
	assert.ErrorIs(t, err, ErrUniqueGetterOrIDGetterExpected)

	getIDEntry1 := &getIDEntryWithOnCreateAndOnUpdate{ID: 1}
	err = orm.Delete(context.Background(), getIDEntry1)
	assert.Nil(t, err)

	err = orm.Delete(context.Background(), getIDEntry1)
	assert.ErrorIs(t, err, ErrNotFound)

	getUniqueEntry1 := &getUniqueEntryWithOnCreateAndOnUpdate{ID1: 1, ID2: 2}
	err = orm.Delete(context.Background(), getUniqueEntry1)
	assert.Nil(t, err)

	err = orm.Delete(context.Background(), getUniqueEntry1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func testGetDBWrapper(t *testing.T, orm ORM) {
	assert.NotNil(t, orm.GetDBWrapper())
}

func testWithTX(t *testing.T, orm ORM) {
	rollbackErr := errors.New("error to trigger rollback")

	err := orm.WithTx(func(o ORM) error {
		entry := &getIDEntryWithOnCreateAndOnUpdate{ID: 1}
		if err := o.GetWithXLock(context.Background(), entry); err != nil {
			return err
		}

		if err := o.Update(context.Background(), entry); err != nil {
			return err
		}

		return rollbackErr
	})
	assert.ErrorIs(t, err, rollbackErr)

	entry := &getIDEntryWithOnCreateAndOnUpdate{ID: 1}
	err = orm.Get(context.Background(), entry)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            1,
		StringCol:     "value 1",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, entry)

	err = orm.WithTx(func(o1 ORM) error {
		return o1.WithTx(func(o2 ORM) error {
			return o2.WithTx(func(o3 ORM) error {
				entry := &getIDEntryWithOnCreateAndOnUpdate{ID: 1}
				if err := o3.GetWithXLock(context.Background(), entry); err != nil {
					return err
				}

				if err := o3.Update(context.Background(), entry); err != nil {
					return err
				}

				return rollbackErr
			})
		})
	})
	assert.ErrorIs(t, err, rollbackErr)

	entry = &getIDEntryWithOnCreateAndOnUpdate{ID: 1}
	err = orm.Get(context.Background(), entry)
	assert.Nil(t, err)
	assert.Equal(t, &getIDEntryWithOnCreateAndOnUpdate{
		ID:            1,
		StringCol:     "value 1",
		OnCreateCount: 1,
		OnUpdateCount: 0,
	}, entry)
}
