package sq_test

import (
	"database/sql"
	sq "github.com/goclub/sql"
)

// 当有一张 user 表
// id	name	age	created_at	updated_at	deleted_at
// 7d96ba00-f788-4b2c-86d6-d71d3b41c903	nimo	18	2021-01-06 03:37:36	2021-01-06 03:38:01	NULL
// 根据表的信息写出如下类型

// 定义符合 sq.Tabler 接口的结构体
type TableUser struct {
	// 通过 goclub/sql 提供的 SoftDeleteDeletedAt 表明该表存在软删字段，还可以使用 sq.SoftDeleteIsDeleted  sq.SoftDeleteDeleteTime
	// 它们的功能是让 TableUser 支持 SoftDeleteWhere() SoftDeleteSet() 方法
	sq.SoftDeletedAt
}
// 通过 TableName() 配置表名
func (*TableUser) TableName() string {return "user"}
// 给 user 表的 id 字段增加类型可减少代码中传错 id 的错误
type IDUser string
// 定义符合 sq.Model 接口的结构体
type User struct {
	ID IDUser `db:"id"`
	Name string `db:"name"`
	Age int `db:"age"`
	// CreatedAtUpdatedAt 表明表是支持 created_at 和 updated_at 字段的，还可以使用 sq.CreateTimeUpdateTime sq.GMTCreateGMTUpdate
	sq.CreatedAtUpdatedAt
	// 通过组合 TableUser 让 User 支持 TableName() SoftDeleteWhere() SoftDeleteSet() 等方法
	TableUser
	// 每个 Model 都应该具有生命周期触发函数 BeforeCreate() AfterCreate() BeforeUpdate() AfterUpdate() 方法
	// 通过 sq.DefaultLifeCycle 可配置默认的生命周期触发函数
	sq.DefaultLifeCycle
}
func (u User) PrimaryKey() []sq.Condition {
	return sq.And(u.Column().ID, sq.Equal(u.ID))
}
// 因为 user 表的 id 字段是 uuid，所以在 User 的 BeforeCreate 生命周期去创建 id
func (u *User) BeforeCreate() error {
	if len(u.ID) == 0 {
		u.ID = IDUser(sq.UUID())
	}
	return nil
}
// 为了避免在代码中重复的写 "id" "name" "age" 等字符串时候写错单词导致的错误，实现 Column 方法避免出错。
func (User) Column () (col struct{
	ID sq.Column
	Name sq.Column
	Age sq.Column
}) {
	col.ID = "id"
	col.Name = "name"
	col.Age = "age"
	return
}



type UserWithAddress struct {
	UserID IDUser `db:"user.id"`
	Name string `db:"user.name"`
	Age int `db:"user.age"`
	Address sql.NullString `db:"user_address.address"`
}
func (UserWithAddress) SoftDeleteWhere() (sq.Raw) {return sq.Raw{"`user`.`deleted_at` IS NULL AND `user_address`.`deleted_at` IS NULL", nil}}
func (*UserWithAddress) TableName() string {return "user"}
func (UserWithAddress) RelationJoin() []sq.Join {
	return []sq.Join{
		{
			Type: 	  	   sq.LeftJoin,
			TableName:	   "user_address",
			On:"`user`.`id` = `user_address`.`user_id`",
		},
	}
}

type TableUserAddress struct {
	// 通过 goclub/sql 提供的 SoftDeleteDeletedAt 表明该表存在软删字段，还可以使用 sq.SoftDeleteIsDeleted  sq.SoftDeleteDeleteTime
	// 它们的功能是让 TableUser 支持 SoftDeleteWhere() SoftDeleteSet() 方法
	sq.SoftDeletedAt
}
type UserAddress struct {
	UserID IDUser `db:"user_id"`
	Address string `db:"address"`
	// CreatedAtUpdatedAt 表明表是支持 created_at 和 updated_at 字段的，还可以使用 sq.CreateTimeUpdateTime sq.GMTCreateGMTUpdate
	sq.CreatedAtUpdatedAt
	// 通过组合 TableUser 让 User 支持 TableName() SoftDeleteWhere() SoftDeleteSet() 等方法
	TableUserAddress
	// 每个 Model 都应该具有生命周期触发函数 BeforeCreate() AfterCreate() BeforeUpdate() AfterUpdate() 方法
	// 通过 sq.DefaultLifeCycle 可配置默认的生命周期触发函数
	sq.DefaultLifeCycle
}
func (UserAddress) TableName() string {
	return "user_address"
}
func (UserWithAddress) Column () (col struct{
	UserID sq.Column
	Name sq.Column
	Age sq.Column
	Address sq.Column
}) {
	col.UserID = "user.id"
	col.Name = "user.name"
	col.Age = "user.age"
	col.Address = "user_address.user_id"
	return
}

