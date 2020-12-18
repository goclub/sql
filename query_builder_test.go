package sq_test

import (
	sq "github.com/goclub/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
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
	assert.Equal(t, "SELECT * FROM `user`", query)
	assert.Equal(t, []interface{}(nil), values)
}
func (suite TestQBSuite) TestTableRaw() {
	t := suite.T()
	qb := sq.QB{
		TableRaw: func() (query string, values []interface{}) {
			return "SELECT * FROM `user` WHERE `name` like ?", []interface{}{"%tableRaw%"}
		},
	}
	query, values := qb.SQLSelect()
	assert.Equal(t, "SELECT * FROM SELECT * FROM `user` WHERE `name` like ?", query)
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