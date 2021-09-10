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
		From: &User{},
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}(nil), values)
}
func (suite TestQBSuite) TestTableRaw() {
	t := suite.T()
	{
		qb := sq.QB{
			SelectRaw: []sq.Raw{{Query: "*"}},
			FromRaw: sq.FromRaw{
				TableName:       sq.Raw{"(SELECT * FROM `user` WHERE `name` like ?) as user", []interface{}{"%tableRaw%"}},
				SoftDeleteWhere: sq.Raw{},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT * FROM (SELECT * FROM `user` WHERE `name` like ?) as user", query)
		assert.Equal(t, []interface{}{"%tableRaw%"}, values)
	}
	{
		qb := sq.QB{
			FromRaw: sq.FromRaw{
				TableName:       sq.Raw{"(SELECT * FROM `user` WHERE `name` like ?) as user", []interface{}{"%tableRaw%"}},
				SoftDeleteWhere: sq.Raw{},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "goclub/sql: if qb.SelectRaw is nil or qb.Form is nil then qb.Select can not be nil or empty slice", query)
		assert.Equal(t, []interface{}(nil), values)
	}

}
func (suite TestQBSuite) TestDisableSoftDelete() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			DisableSoftDelete: true,
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user`", query)
		assert.Equal(t, []interface{}(nil), values)
	}
	{
		qb := sq.QB{
			From: &User{},
			DisableSoftDelete: false,
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}(nil), values)
	}
}
func (suite TestQBSuite) TestUnionTable() {
	t := suite.T()
	{
		where := sq.And("age", sq.GtInt(18))
		qb := sq.QB{
			UnionTable: sq.UnionTable{
				Tables:    []sq.QB{
					{
						From:&User{},
						Where: where,
					},
					{
						From:&User{},
						Where: where,
					},
				},
				UnionAll: true,
			},
			Where: []sq.Condition{
				{"id", sq.Equal(1)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "(SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL) UNION ALL (SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL) WHERE `id` = ?", query)
		assert.Equal(t, []interface{}{18, 18, 1}, values)
	}
	{
		where := sq.And("age", sq.GtInt(18))
		qb := sq.QB{
			UnionTable: sq.UnionTable{
				Tables:    []sq.QB{
					{
						From:&User{},
						Where: where,
					},
					{
						From:&User{},
						Where: where,
					},
				},
				UnionAll: false,
			},
			Where: []sq.Condition{
				{"id", sq.Equal(1)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "(SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL) UNION (SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL) WHERE `id` = ?", query)
		assert.Equal(t, []interface{}{18, 18, 1}, values)
	}
}
func (suite TestQBSuite) TestSelect() {
	t := suite.T()
	{
		qb := sq.QB{
			FromRaw: sq.FromRaw{
				TableName: sq.Raw{"user", nil},
				SoftDeleteWhere: sq.Raw{},
			},
			SelectRaw: []sq.Raw{{Query: "*"}},
			Where: []sq.Condition{
				{"id", sq.Equal(1)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT * FROM user WHERE `id` = ?", query)
		assert.Equal(t, []interface{}{1}, values)
	}
	{
		qb := sq.QB{
			FromRaw: sq.FromRaw{
				TableName: sq.Raw{"user", nil},
				SoftDeleteWhere: sq.Raw{},
			},
			Select: nil,
			Where: []sq.Condition{
				{"id", sq.Equal(1)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "goclub/sql: if qb.SelectRaw is nil or qb.Form is nil then qb.Select can not be nil or empty slice", query)
		assert.Nil(t, values)
	}
	{
		qb := sq.QB{
			FromRaw: sq.FromRaw{
				TableName:       sq.Raw{"user", nil},
				SoftDeleteWhere: sq.Raw{},
			},
			Select: []sq.Column{"name"},
			Where: []sq.Condition{
				{"id", sq.Equal(1)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `name` FROM user WHERE `id` = ?", query)
		assert.Equal(t, []interface{}{1}, values)
	}
}
func (suite TestQBSuite) TestSelectRaw() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		// Select 会被忽略 优先使用 SelectRaw
		Select: []sq.Column{"name"},
		SelectRaw: []sq.Raw{
			sq.Raw{"count(*) as count",nil},
		},
		Where: []sq.Condition{
			{"id", sq.Equal(1)},
		},
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	assert.Equal(t, "SELECT count(*) as count FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{1}, values)
}
func (suite TestQBSuite) TestSelectColumnHasDot() {
	t := suite.T()
	qb := sq.QB{
		FromRaw: sq.FromRaw{
			TableName:       sq.Raw{"user as u", nil},
			SoftDeleteWhere: sq.Raw{},
		},
		Select: []sq.Column{`u.name`,},
		Where: []sq.Condition{
			{"id", sq.Equal(1)},
		},
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	assert.Equal(t, "SELECT `u`.`name` AS 'u.name' FROM user as u WHERE `id` = ?", query)
	assert.Equal(t, []interface{}{1}, values)
}
func (suite TestQBSuite) TestIndex() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		Index: "USE INDEX(PRIMARY)",
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` USE INDEX(PRIMARY) WHERE `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}(nil), values)
}

func (suite TestQBSuite) TestWhere() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				{"name", sq.Equal("nimo")},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, sq.ToConditions(qb.Where), sq.And("name", sq.Equal("nimo")))
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo"}, values)
	}
}

func (suite TestQBSuite) TestWhereOR() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			WhereOR: [][]sq.Condition{
				{{"name", sq.Equal("nimo")}},
				{{"name", sq.Equal("nico")}},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE (`name` = ?) OR (`name` = ?) AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo", "nico"}, values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Select: []sq.Column{"id"},
			// WHERE (`name` LIKE ? OR `mobile` LIKE ?) AND `role_id` = ?
			Where: sq.
				OrGroup(
					sq.Condition{"name", sq.Like("nimo")},
					sq.Condition{"mobile", sq.Like("13611112222")},
				).
				And("role_id", sq.Equal("1")),
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		ands := []sq.Condition{
			{
				"", sq.OP{
				OrGroup: []sq.Condition{
					{"name", sq.Like("nimo")},
					{"mobile", sq.Like("13611112222")},
				},
			},
			},
			{"role_id", sq.Equal("1")},
		}
		assert.Equal(t, ands, qb.Where)
		assert.Equal(t, "SELECT `id` FROM `user` WHERE (`name` LIKE ? OR `mobile` LIKE ?) AND `role_id` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"%nimo%", "%13611112222%", "1",}, values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Select: []sq.Column{"id"},
			// WHERE (`name` LIKE ? OR `mobile` LIKE ?) AND `role_id` = ?
			Where: sq.
				OrGroup(
					sq.Condition{"name", sq.Ignore(false, sq.Like("nimo"))},
					sq.Condition{"mobile", sq.Ignore(true, sq.Like("13611112222"))},
			).
			And("role_id", sq.Equal("1")),
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id` FROM `user` WHERE (`name` LIKE ?) AND `role_id` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"%nimo%", "1",}, values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Select: []sq.Column{"id"},
			// WHERE (`name` LIKE ? OR `mobile` LIKE ?) AND `role_id` = ?
			Where: sq.
				OrGroup(
					sq.Condition{"name", sq.Ignore(true, sq.Like("nimo"))},
					sq.Condition{"mobile", sq.Ignore(false, sq.Like("13611112222"))},
				).
				And("role_id", sq.Equal("1")),
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id` FROM `user` WHERE (`mobile` LIKE ?) AND `role_id` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"%13611112222%", "1",}, values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Select: []sq.Column{"id"},
			// WHERE (`name` LIKE ? OR `mobile` LIKE ?) AND `role_id` = ?
			Where: sq.
				OrGroup(
					sq.Condition{"name", sq.Ignore(true, sq.Like("nimo"))},
					sq.Condition{"mobile", sq.Ignore(true, sq.Like("13611112222"))},
				).
				And("role_id", sq.Equal("1")),
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id` FROM `user` WHERE `role_id` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"1",}, values)
	}
}

func (suite TestQBSuite) TestWhereRaw() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			WhereRaw: sq.Raw{"`name` = ? AND `age` = ?", []interface{}{"nimo", 1}},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` = ? AND `age` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo", 1}, values)
	}
	{}
}
func (suite TestQBSuite) TestWhereOPRaw() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				sq.ConditionRaw("`name` = `cname`", nil),
				sq.ConditionRaw("`age` = ?", []interface{}{1}),
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` = `cname` AND `age` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{1}, values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				sq.ConditionRaw("`name` = ?", []interface{}{"nimo"}),
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo"}, values)
	}
}
func (suite TestQBSuite) TestWhereSubQuery() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				{"id", sq.SubQuery("IN", sq.QB{
					From: &User{},
					Select: []sq.Column{"id"},
				})},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` IN (SELECT `id` FROM `user` WHERE `deleted_at` IS NULL) AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}(nil), values)
	}
}
func (suite TestQBSuite) TestWhereAndTwoCondition() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				{"name", sq.Equal("nimo")},
				{"age", sq.Equal(18)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		ands := []sq.Condition(
			sq.And("name", sq.Equal("nimo")).And("age", sq.Equal(18)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` = ? AND `age` = ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{"nimo", 18}, values)
	}
}
func (suite TestQBSuite) TestWhereOPGtIntLtInt() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				{"age", sq.GtInt(18)},
				{"age", sq.LtInt(19)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtInt(18)).And("age", sq.LtInt(19)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `age` < ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{18, 19}, values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				{"age", sq.GtOrEqualInt(18)},
				{"age", sq.LtOrEqualInt(19)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtOrEqualInt(18)).And("age", sq.LtOrEqualInt(19)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` >= ? AND `age` <= ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{18, 19}, values)
	}
}
func (suite TestQBSuite) TestWhereOPGtFloatLtFloat() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				{"age", sq.GtFloat(18.11)},
				{"age", sq.LtFloat(19.22)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtFloat(18.11)).And("age", sq.LtFloat(19.22)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `age` < ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{18.11, 19.22}, values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				{"age", sq.GtOrEqualFloat(18.11)},
				{"age", sq.LtOrEqualFloat(19.22)},
			},
		}
		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
		ands := []sq.Condition(
			sq.And("age", sq.GtOrEqualFloat(18.11)).And("age", sq.LtOrEqualFloat(19.22)),
		)
		assert.Equal(t, qb.Where, ands)
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` >= ? AND `age` <= ? AND `deleted_at` IS NULL", query)
		assert.Equal(t, []interface{}{18.11, 19.22}, values)
	}
}
// func (suite TestQBSuite) TestWhereOPGtTimeLtTime() {
// 	t := suite.T()
// 	startTime := time.Date(2020,11,11,22,22,22,0, time.UTC)
// 	endTime := time.Date(2020,11,11,22,22,22,0, time.UTC)
//
// 	{
// 		qb := sq.QB{
// 			From: &User{},
// 			Where: []sq.Condition{
// 				{"age", sq.GtTime(startTime)},
// 				{"age", sq.LtTime(endTime)},
// 			},
// 		}
// 		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
// 		ands := []sq.Condition(
// 			sq.And("age", sq.GtTime(startTime)).And("age", sq.LtTime(endTime)),
// 		)
// 		assert.Equal(t, qb.Where, ands)
// 		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `age` < ? AND `deleted_at` IS NULL", query)
// 		assert.Equal(t, []interface{}{startTime, endTime}, values)
// 	}
// 	{
// 		qb := sq.QB{
// 			From: &User{},
// 			Where: []sq.Condition{
// 				{"age", sq.GtOrEqualTime(startTime)},
// 				{"age", sq.LtOrEqualTime(endTime)},
// 			},
// 		}
// 		raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
// 		ands := []sq.Condition(
// 			sq.And("age", sq.GtOrEqualTime(startTime)).And("age", sq.LtOrEqualTime(endTime)),
// 		)
// 		assert.Equal(t, qb.Where, ands)
// 		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` >= ? AND `age` <= ? AND `deleted_at` IS NULL", query)
// 		assert.Equal(t, []interface{}{startTime, endTime}, values)
// 	}
// }

func (suite TestQBSuite) TestWhereEqualAndNotEqual() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		Where: []sq.Condition{
			{"name", sq.Equal("nimo")},
			{"book", sq.NotEqual("abc")},
		},
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	ands := []sq.Condition(
		sq.And("name", sq.Equal("nimo")).And("book", sq.NotEqual("abc")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` = ? AND `book` <> ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"nimo", "abc"}, values)
}

func (suite TestQBSuite) TestWhereLike() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		Where: []sq.Condition{
			{"name", sq.Like("nimo")},
		},
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	ands := []sq.Condition(
		sq.And("name", sq.Like("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"%nimo%"}, values)
}

func (suite TestQBSuite) TestWhereLikeLeft() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		Where: []sq.Condition{
			{"name", sq.LikeLeft("nimo")},
		},
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	ands := []sq.Condition(
		sq.And("name", sq.LikeLeft("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"nimo%"}, values)
}
func (suite TestQBSuite) TestWhereLikeRight() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		Where: []sq.Condition{
			{"name", sq.LikeRight("nimo")},
		},
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	ands := []sq.Condition(
		sq.And("name", sq.LikeRight("nimo")),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"%nimo"}, values)
}


func (suite TestQBSuite) TestWhereIn() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		Where: []sq.Condition{
			{"id", sq.In([]string{"a","b"})},
		},
	}
	raw := qb.SQLSelect(); query, values :=  raw.Query, raw.Values
	ands := []sq.Condition(
		sq.And("id", sq.In([]string{"a","b"})),
	)
	assert.Equal(t, qb.Where, ands)
	assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` IN (?,?) AND `deleted_at` IS NULL", query)
	assert.Equal(t, []interface{}{"a","b"}, values)
}
func (suite TestQBSuite) TestWhereIgnore() {
	t := suite.T()
	test := func (searchName string, query string, values []interface{}) {
		qb := sq.QB{
			From: &User{},
			Select: []sq.Column{"id"},
			Where: sq.And("name", sq.Ignore(searchName == "", sq.Equal(searchName))),
			Reviews: []string{
				"SELECT `id` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL",
				"SELECT `id` FROM `user` WHERE `deleted_at` IS NULL",
			},
			SQLChecker: sq.DefaultSQLChecker{},
		}
		raw := qb.SQLSelect()
		assert.Equal(t, query, raw.Query)
		assert.Equal(t, values, raw.Values)
	}
	test("", "SELECT `id` FROM `user` WHERE `deleted_at` IS NULL", []interface{}(nil))
	test("nimo", "SELECT `id` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL", []interface{}{"nimo"})
	{
		raw := sq.QB{
			From: &User{},
			Where: sq.And("name", sq.Ignore(true, sq.Equal("nimo"))),
			DisableSoftDelete: true,
		}.SQLSelect()
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user`", raw.Query)
		assert.Equal(t, []interface{}(nil), raw.Values)
	}
	{
		raw := sq.QB{
			From: &User{},
			Where: sq.And("name", sq.Ignore(true, sq.Equal("nimo"))).And("age", sq.Ignore(true, sq.Equal(1))),
			DisableSoftDelete: true,
		}.SQLSelect()
		assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user`", raw.Query)
		assert.Equal(t, []interface{}(nil), raw.Values)
	}

}
func (suite TestQBSuite) TestInPanic() {
	t := suite.T()
	var panicValue interface{}
	func() {
		defer func() {
			panicValue = recover()
		}()
		qb := sq.QB{
			From: &User{},
			Where: []sq.Condition{
				{"id", sq.In("a")},
			},
		}
		_ = qb.SQLSelect()
	}()

	assert.Equal(t, panicValue.(error).Error(), "sq.In(string) slice must be slice")
}

func (suite TestQBSuite) TestLimit() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Limit: 1,
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL LIMIT ?")
		assert.Equal(t, []interface{}{1}, raw.Values)
	}
	{
		qb := sq.QB{
			From: &User{},
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL")
		assert.Equal(t, []interface{}(nil), raw.Values)
	}
}
func (suite TestQBSuite) TestOffset() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Offset: 100,
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL OFFSET ?")
		assert.Equal(t, []interface{}{100}, raw.Values)
	}
	{
		qb := sq.QB{
			From: &User{},
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL")
		assert.Equal(t, []interface{}(nil), raw.Values)
	}
}
func (suite TestQBSuite) TestLimitOffset() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Limit:2,
			Offset: 100,
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL LIMIT ? OFFSET ?")
		assert.Equal(t, []interface{}{2, 100}, raw.Values)
	}
}
func (suite TestQBSuite) TestLock() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Lock: sq.FORSHARE,
			Limit: 1,
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL LIMIT ? FOR SHARE")
		assert.Equal(t, []interface{}{1}, raw.Values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Lock: sq.FORUPDATE,
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL FOR UPDATE")
		assert.Equal(t, []interface{}(nil), raw.Values)
	}
}
func (suite TestQBSuite) TestJoin() {
	t := suite.T()
	uaCol := UserWithAddress{}.Column()
	qb := sq.QB{
		FromRaw: sq.FromRaw{
			TableName:       sq.Raw{"user", nil},
			SoftDeleteWhere: sq.Raw{"`user`.`deleted_at` is NULL AND user_address`.`deleted_at` is NULL", nil},
		},
		Select: []sq.Column{"user.id", "user_address.address"},
		Where: sq.And(uaCol.UserID, sq.Equal(1)),
		Join: []sq.Join{
			{
				Type: sq.LeftJoin,
				TableName: "`user_address`",
				On: "`user`.`id` = `user_address`.`user_id`",
			},
		},
	}
	raw := qb.SQLSelect()
	assert.Equal(t, raw.Query, "SELECT `user`.`id` AS 'user.id', `user_address`.`address` AS 'user_address.address' FROM user LEFT JOIN ``user_address`` ON `user`.`id` = `user_address`.`user_id` WHERE `user`.`id` = ? AND `user`.`deleted_at` is NULL AND user_address`.`deleted_at` is NULL")
	assert.Equal(t, []interface{}{1}, raw.Values)
}
func (suite TestQBSuite) TestJoinType() {
	t := suite.T()
	assert.Equal(t, sq.LeftJoin.String(), "LEFT JOIN")
	assert.Equal(t, sq.LeftJoin.String(), string(sq.LeftJoin))
}
func (suite TestQBSuite) TestStatement() {
	t := suite.T()
	assert.Equal(t, sq.StatementSelect.String(), "SELECT")
	assert.Equal(t, sq.StatementSelect.String(), string(sq.StatementSelect))
}



func (suite TestQBSuite) TestOrderBy() {
	t := suite.T()
	{
		qb := sq.QB{
			From: &User{},
			Limit: 2,
			Offset:10,
			OrderBy: []sq.OrderBy{{Column: "name"}},
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL ORDER BY `name` ASC LIMIT ? OFFSET ?")
		assert.Equal(t, []interface{}{2,10}, raw.Values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Limit: 2,
			Offset:10,
			OrderBy: []sq.OrderBy{{"name", sq.DESC}},
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL ORDER BY `name` DESC LIMIT ? OFFSET ?")
		assert.Equal(t, []interface{}{2,10}, raw.Values)
	}
	{
		qb := sq.QB{
			From: &User{},
			Limit: 2,
			Offset:10,
			OrderBy: []sq.OrderBy{{"name", sq.DESC,}, {"age", sq.ASC}},
		}
		raw := qb.SQLSelect()
		assert.Equal(t, raw.Query, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL ORDER BY `name` DESC, `age` ASC LIMIT ? OFFSET ?")
		assert.Equal(t, []interface{}{2,10}, raw.Values)
	}
}
func (suite TestQBSuite) TestUnsafeDelete() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
	}
	raw := qb.SQLDelete()
	assert.Equal(t, "goclub/sql:(MAYBE_FORGET_WHERE)", raw.Query)
	assert.Equal(t, []interface{}(nil), raw.Values)
}
func (suite TestQBSuite) TestHaving() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		SelectRaw: []sq.Raw{
			{"`name`", nil},
			{"count(*) AS count", nil},
		},
		GroupBy: []sq.Column{"name"},
		Having: sq.And("count", sq.GtInt(1)),
	}
	raw := qb.SQLSelect()
	assert.Equal(t, "SELECT `name`, count(*) AS count FROM `user` WHERE `deleted_at` IS NULL GROUP BY `name` HAVING `count` > ?", raw.Query)
	assert.Equal(t, []interface{}{1}, raw.Values)
}
func (suite TestQBSuite) TestInsert() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		UseInsertIgnoreInto: true,
		Insert: []sq.Insert{
			{"name", "nimoc"},
		},
	}
	raw := qb.SQLInsert()
	assert.Equal(t, "INSERT IGNORE INTO `user` (`name`) VALUES (?)", raw.Query)
	assert.Equal(t, []interface{}{"nimoc"}, raw.Values)
}

func (suite TestQBSuite) TestInsertMultiple() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		InsertMultiple: sq.InsertMultiple{
			Column: []sq.Column{"name", "age"},
			Values: [][]interface{}{
				{"nimo",18},
				{"tim", 28},
			},
		},
	}
	raw := qb.SQLInsert()
	assert.Equal(t, "INSERT INTO `user` (`name`,`age`) VALUES (?,?),(?,?)", raw.Query)
	assert.Equal(t, []interface{}{"nimo",18,"tim",28}, raw.Values)
}

func (suite TestQBSuite) TestUpdate() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		UseUpdateIgnore: true,
		Update: sq.Set("age", 2),
		Where:  sq.And("id", sq.Equal(1)),
	}
	raw := qb.SQLUpdate()
	assert.Equal(t, "UPDATE IGNORE `user` SET `age`= ? WHERE `id` = ? AND `deleted_at` IS NULL", raw.Query)
	assert.Equal(t, []interface{}{2, 1}, raw.Values)
}
func (suite TestQBSuite) TestDebug() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		Where:  sq.
			And("id", sq.Equal(1)).
			And("date", sq.Equal(time.Now())),
		Debug: true,
	}
	qb.SQLSelect()
	_=t
}
func (suite TestQBSuite) TestSet() {
	t := suite.T()
	{
		update := sq.Set("id", sq.Equal(1))
		assert.Equal(t, update[0].Value, 1)
	}
	{
		update := sq.Set("id", sq.Equal(2))
		assert.Equal(t, update[0].Value, 2)
	}
}


func (suite TestQBSuite) TestGroupByOrderBy() {
	t := suite.T()
	qb := sq.QB{
		From: &User{},
		GroupBy: []sq.Column{"date"},
		OrderBy: []sq.OrderBy{
			{"date", sq.DESC,},
		},
	}
	raw := qb.SQLSelect()
	assert.Equal(t, "SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `deleted_at` IS NULL GROUP BY `date` ORDER BY `date` DESC", raw.Query)
}