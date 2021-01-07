package sq_test

import (
	sq "github.com/goclub/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)
func TestQB(t *testing.T) {
	suite.Run(t, new(TestQBSuite))
}
type TestQBSuite struct {
	suite.Suite
}
func (suite TestQBSuite) TestTable() {
	t := suite.T()
	qb := sq.QB{
		Table: User{},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	assert.Equal(t, "SELECT * FROM `user` WHERE `is_deleted` = 0", query)
	assert.Equal(t, []interface{}(nil), values)
}
func (suite TestQBSuite) TestTableRaw() {
	t := suite.T()
	qb := sq.QB{
		TableRaw: sq.QueryValues{"(SELECT * FROM `user` WHERE `name` like ?) as user", []interface{}{"%tableRaw%"}},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	assert.Equal(t, "SELECT * FROM (SELECT * FROM `user` WHERE `name` like ?) as user", query)
	assert.Equal(t, []interface{}{"%tableRaw%"}, values)
}


func (suite TestQBSuite) TestIndex() {
	t := suite.T()
	qb := sq.QB{
		Table: User{},
		Index: "USE INDEX(PRIMARY)",
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	assert.Equal(t, "SELECT * FROM `user` USE INDEX(PRIMARY) WHERE `is_deleted` = 0", query)
	assert.Equal(t, []interface{}(nil), values)
}

func (suite TestQBSuite) TestWhere() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"name", sq.Equal("nimo")},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		assert.Equal(t, sq.ToConditions(qb.Where), sq.And("name", sq.Equal("nimo")))
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{"nimo"}, values)
	}
}
func (suite TestQBSuite) TestWhereOR() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: User{},
			WhereOR: [][]sq.Condition{
				{{"name", sq.Equal("nimo")}},
				{{"name", sq.Equal("nico")}},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		assert.Equal(t, "SELECT * FROM `user` WHERE (`name` = ?) OR (`name` = ?) AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{"nimo", "nico"}, values)
	}
}

func (suite TestQBSuite) TestWhereRaw() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: User{},
			WhereRaw: func() sq.QueryValues {
				return sq.QueryValues{"`name` = ? AND `age` = ?", []interface{}{"nimo", 1}}
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `age` = ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{"nimo", 1}, values)
	}
}
func (suite TestQBSuite) TestWhereOPRaw() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"", sq.Raw("`name` = `cname`")},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = `cname` AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}(nil), values)
	}
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"", sq.Raw("`name` = ?", "nimo")},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{"nimo"}, values)
	}
}
func (suite TestQBSuite) TestWhereSubQuery() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"id", sq.SubQuery("IN", sq.QB{
					Table: User{},
					Select: []sq.Column{"id"},
				})},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		assert.Equal(t, "SELECT * FROM `user` WHERE `id` IN (SELECT `id` FROM `user` WHERE `is_deleted` = 0) AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}(nil), values)
	}
}
func (suite TestQBSuite) TestWhereAndTwoCondition() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"name", sq.Equal("nimo")},
				{"age", sq.Equal(18)},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		ands := []sq.Condition(
			sq.And("name", sq.Equal("nimo")).And("age", sq.Equal(18)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `age` = ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{"nimo", 18}, values)
	}
}
func (suite TestQBSuite) TestWhereOPGtIntLtInt() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"age", sq.GtInt(18)},
				{"age", sq.LtInt(19)},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtInt(18)).And("age", sq.LtInt(19)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` > ? AND `age` < ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{18, 19}, values)
	}
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"age", sq.GtOrEqualInt(18)},
				{"age", sq.LtOrEqualInt(19)},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtOrEqualInt(18)).And("age", sq.LtOrEqualInt(19)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` >= ? AND `age` <= ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{18, 19}, values)
	}
}
func (suite TestQBSuite) TestWhereOPGtFloatLtFloat() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"age", sq.GtFloat(18.11)},
				{"age", sq.LtFloat(19.22)},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtFloat(18.11)).And("age", sq.LtFloat(19.22)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` > ? AND `age` < ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{18.11, 19.22}, values)
	}
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"age", sq.GtOrEqualFloat(18.11)},
				{"age", sq.LtOrEqualFloat(19.22)},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtOrEqualFloat(18.11)).And("age", sq.LtOrEqualFloat(19.22)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` >= ? AND `age` <= ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{18.11, 19.22}, values)
	}
}
func (suite TestQBSuite) TestWhereOPGtTimeLtTime() {
	t := suite.T()
	startTime := time.Date(2020,11,11,22,22,22,0, time.UTC)
	endTime := time.Date(2020,11,11,22,22,22,0, time.UTC)

	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"age", sq.GtTime(startTime)},
				{"age", sq.LtTime(endTime)},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtTime(startTime)).And("age", sq.LtTime(endTime)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` > ? AND `age` < ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{startTime, endTime}, values)
	}
	{
		qb := sq.QB{
			Table: User{},
			Where: []sq.Condition{
				{"age", sq.GtOrEqualTime(startTime)},
				{"age", sq.LtOrEqualTime(endTime)},
			},
		}
		qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtOrEqualTime(startTime)).And("age", sq.LtOrEqualTime(endTime)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` >= ? AND `age` <= ? AND `is_deleted` = 0", query)
		assert.Equal(t, []interface{}{startTime, endTime}, values)
	}
}

func (suite TestQBSuite) TestWhereEqualAndNotEqual() {
	t := suite.T()
	qb := sq.QB{
		Table: User{},
		Where: []sq.Condition{
			{"name", sq.Equal("nimo")},
			{"book", sq.NotEqual("abc")},
		},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	ands := []sq.Condition(
		sq.And("name", sq.Equal("nimo")).And("book", sq.NotEqual("abc")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `book` <> ? AND `is_deleted` = 0", query)
	assert.Equal(t, []interface{}{"nimo", "abc"}, values)
}

func (suite TestQBSuite) TestLike() {
	t := suite.T()
	qb := sq.QB{
		Table: User{},
		Where: []sq.Condition{
			{"name", sq.Like("nimo")},
		},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	ands := []sq.Condition(
		sq.And("name", sq.Like("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `name` LIKE ? AND `is_deleted` = 0", query)
	assert.Equal(t, []interface{}{"%nimo%"}, values)
}

func (suite TestQBSuite) TestLikeLeft() {
	t := suite.T()
	qb := sq.QB{
		Table: User{},
		Where: []sq.Condition{
			{"name", sq.LikeLeft("nimo")},
		},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	ands := []sq.Condition(
		sq.And("name", sq.LikeLeft("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `name` LIKE ? AND `is_deleted` = 0", query)
	assert.Equal(t, []interface{}{"nimo%"}, values)
}
func (suite TestQBSuite) TestLikeRight() {
	t := suite.T()
	qb := sq.QB{
		Table: User{},
		Where: []sq.Condition{
			{"name", sq.LikeRight("nimo")},
		},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	ands := []sq.Condition(
		sq.And("name", sq.LikeRight("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `name` LIKE ? AND `is_deleted` = 0", query)
	assert.Equal(t, []interface{}{"%nimo"}, values)
}


func (suite TestQBSuite) TestIn() {
	t := suite.T()
	qb := sq.QB{
		Table: User{},
		Where: []sq.Condition{
			{"id", sq.In([]string{"a","b"})},
		},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	ands := []sq.Condition(
		sq.And("id", sq.In([]string{"a","b"})),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `id` IN (?, ?) AND `is_deleted` = 0", query)
	assert.Equal(t, []interface{}{"a","b"}, values)
}
func (suite TestQBSuite) TestInPanic() {
	t := suite.T()
	qb := sq.QB{
		Table: User{},
		Where: []sq.Condition{
			{"id", sq.In("a")},
		},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	ands := []sq.Condition(
		sq.And("id", sq.In([]string{"a","b"})),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `id` IN (?, ?) AND `is_deleted` = 0", query)
	assert.Equal(t, []interface{}{"a","b"}, values)
}

func (suite TestQBSuite) TestUnionTable() {
	t := suite.T()
	where := sq.And("age", sq.GtInt(18))
	qb := sq.QB{
		Union: sq.Union{
			Tables:    []sq.QB{
				{
					Table:User{},
					Where: where,
				},
				{
					Table:User{},
					Where: where,
				},
			},
			UnionAll: true,
		},
		Where: []sq.Condition{
			{"id", sq.Equal(1)},
		},
	}
	qv := qb.SQLSelect(); query, values :=  qv.Query, qv.Values
	assert.Equal(t, "(SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL) UNION ALL (SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL) WHERE `id` = ?", query)
	assert.Equal(t, []interface{}{18, 18, 1}, values)
}