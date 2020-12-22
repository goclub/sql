package sq_test

import (
	sq "github.com/goclub/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)
type TestQBSuite struct {
	suite.Suite
}
func (suite TestQBSuite) TestTable() {
	t := suite.T()
	qb := sq.QB{
		Table: "user",
	}
	query, values := qb.SQLSelect()
	assert.Equal(t, "SELECT * FROM `user` WHERE `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}(nil), values)
}
func (suite TestQBSuite) TestTableRaw() {
	t := suite.T()
	qb := sq.QB{
		TableRaw: func() (query string, values []interface{}) {
			return "(SELECT * FROM `user` WHERE `name` like ?) as user", []interface{}{"%tableRaw%"}
		},
	}
	query, values := qb.SQLSelect()
	assert.Equal(t, "SELECT * FROM (SELECT * FROM `user` WHERE `name` like ?) as user WHERE `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"%tableRaw%"}, values)
}
// func (suite TestQBSuite) TestTableRawUserQB() {
// 	t := suite.T()
// 	qb := sq.QB{
// 		TableRaw: sq.QB{
// 			Table: "user",
// 			Where: []sq.Condition{{"name", sq.Like("tableRaw")}},
// 		}.SQLSelect,
// 	}
// 	query, values := qb.SQLSelect()
// 	assert.Equal(t, "SELECT * FROM SELECT * FROM `user` WHERE `name` like ?", query)
// 	assert.Equal(t, []interface{}{"%tableRaw%"}, values)
// }
func TestQB(t *testing.T) {
	suite.Run(t, new(TestQBSuite))
}
func (suite TestQBSuite) TestWhere() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"name", sq.Equal("nimo")},
			},
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, sq.ToConditions(qb.Where), sq.And("name", sq.Equal("nimo")))
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo"}, values)
	}
}
func (suite TestQBSuite) TestWhereOR() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			WhereOR: [][]sq.Condition{
				{{"name", sq.Equal("nimo")}},
				{{"name", sq.Equal("nico")}},
			},
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user` WHERE (`name` = ?) OR (`name` = ?) AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo", "nico"}, values)
	}
}

func (suite TestQBSuite) TestWhereRaw() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			WhereRaw: func() (query string, values []interface{}) {
				return "`name` = ? AND `age` = ?", []interface{}{"nimo", 1}
			},
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `age` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo", 1}, values)
	}
}
func (suite TestQBSuite) TestCustomSoftDelete() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"name", sq.Equal("nimo")},
			},
			SoftDelete: "del_at",
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `del_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo"}, values)
	}
	{
		qb := sq.QB{
			Table: "user",
			SoftDelete: "del_at",
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user` WHERE `del_at` IS NULL", query)
		assert.Equal(t, []interface{}(nil), values)
	}
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"name", sq.Equal("nimo")},
			},
			SoftDelete: sq.DisableSoftDelete,
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ?", query)
		assert.Equal(t, []interface{}{"nimo"}, values)
	}
	{
		qb := sq.QB{
			Table: "user",
			SoftDelete: sq.DisableSoftDelete,
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user`", query)
		assert.Equal(t, []interface{}(nil), values)
	}
}
func (suite TestQBSuite) TestWhereOPRaw() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"", sq.Raw("`name` = `cname`")},
			},
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = `cname` AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}(nil), values)
	}
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"", sq.Raw("`name` = ?", "nimo")},
			},
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo"}, values)
	}
}
func (suite TestQBSuite) TestWhereSubQuery() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"id", sq.SubQuery("IN", sq.QB{
					Table: "user",
					Select: []sq.Column{"id"},
				})},
			},
		}
		query, values := qb.SQLSelect()
		assert.Equal(t, "SELECT * FROM `user` WHERE `id` IN (SELECT `id` FROM `user` WHERE `deleted_at` IS NULL) AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}(nil), values)
	}
}
func (suite TestQBSuite) TestWhereAndTwoCondition() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"name", sq.Equal("nimo")},
				{"age", sq.Equal(18)},
			},
		}
		query, values := qb.SQLSelect()
		ands := []sq.Condition(
			sq.And("name", sq.Equal("nimo")).And("age", sq.Equal(18)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `age` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo", 18}, values)
	}
}
func (suite TestQBSuite) TestWhereOPGtIntLtInt() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"age", sq.GtInt(18)},
				{"age", sq.LtInt(19)},
			},
		}
		query, values := qb.SQLSelect()
		ands := []sq.Condition(
			sq.And("age", sq.GtInt(18)).And("age", sq.LtInt(19)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` > ? AND `age` < ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{18, 19}, values)
	}
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"age", sq.GtOrEqualInt(18)},
				{"age", sq.LtOrEqualInt(19)},
			},
		}
		query, values := qb.SQLSelect()
		ands := []sq.Condition(
			sq.And("age", sq.GtOrEqualInt(18)).And("age", sq.LtOrEqualInt(19)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` >= ? AND `age` <= ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{18, 19}, values)
	}
}
func (suite TestQBSuite) TestWhereOPGtFloatLtFloat() {
	t := suite.T()
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"age", sq.GtFloat(18.11)},
				{"age", sq.LtFloat(19.22)},
			},
		}
		query, values := qb.SQLSelect()
		ands := []sq.Condition(
			sq.And("age", sq.GtFloat(18.11)).And("age", sq.LtFloat(19.22)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` > ? AND `age` < ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{18.11, 19.22}, values)
	}
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"age", sq.GtOrEqualFloat(18.11)},
				{"age", sq.LtOrEqualFloat(19.22)},
			},
		}
		query, values := qb.SQLSelect()
		ands := []sq.Condition(
			sq.And("age", sq.GtOrEqualFloat(18.11)).And("age", sq.LtOrEqualFloat(19.22)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` >= ? AND `age` <= ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{18.11, 19.22}, values)
	}
}
func (suite TestQBSuite) TestWhereOPGtTimeLtTime() {
	t := suite.T()
	startTime := time.Date(2020,11,11,22,22,22,0, time.UTC)
	endTime := time.Date(2020,11,11,22,22,22,0, time.UTC)

	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"age", sq.GtTime(startTime)},
				{"age", sq.LtTime(endTime)},
			},
		}
		query, values := qb.SQLSelect()
		ands := []sq.Condition(
			sq.And("age", sq.GtTime(startTime)).And("age", sq.LtTime(endTime)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` > ? AND `age` < ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{startTime, endTime}, values)
	}
	{
		qb := sq.QB{
			Table: "user",
			Where: []sq.Condition{
				{"age", sq.GtOrEqualTime(startTime)},
				{"age", sq.LtOrEqualTime(endTime)},
			},
		}
		query, values := qb.SQLSelect()
		ands := []sq.Condition(
			sq.And("age", sq.GtOrEqualTime(startTime)).And("age", sq.LtOrEqualTime(endTime)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT * FROM `user` WHERE `age` >= ? AND `age` <= ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{startTime, endTime}, values)
	}
}

func (suite TestQBSuite) TestWhereEqualAndNotEqual() {
	t := suite.T()
	qb := sq.QB{
		Table: "user",
		Where: []sq.Condition{
			{"name", sq.Equal("nimo")},
			{"book", sq.NotEqual("abc")},
		},
	}
	query, values := qb.SQLSelect()
	ands := []sq.Condition(
		sq.And("name", sq.Equal("nimo")).And("book", sq.NotEqual("abc")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `name` = ? AND `book` <> ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"nimo", "abc"}, values)
}

func (suite TestQBSuite) TestLike() {
	t := suite.T()
	qb := sq.QB{
		Table: "user",
		Where: []sq.Condition{
			{"name", sq.Like("nimo")},
		},
	}
	query, values := qb.SQLSelect()
	ands := []sq.Condition(
		sq.And("name", sq.Like("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"%nimo%"}, values)
}

func (suite TestQBSuite) TestLikeLeft() {
	t := suite.T()
	qb := sq.QB{
		Table: "user",
		Where: []sq.Condition{
			{"name", sq.LikeLeft("nimo")},
		},
	}
	query, values := qb.SQLSelect()
	ands := []sq.Condition(
		sq.And("name", sq.LikeLeft("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"nimo%"}, values)
}
func (suite TestQBSuite) TestLikeRight() {
	t := suite.T()
	qb := sq.QB{
		Table: "user",
		Where: []sq.Condition{
			{"name", sq.LikeRight("nimo")},
		},
	}
	query, values := qb.SQLSelect()
	ands := []sq.Condition(
		sq.And("name", sq.LikeRight("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT * FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"%nimo"}, values)
}