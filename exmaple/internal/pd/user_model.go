package pd

import (
	sq "github.com/goclub/sql"
)

// 给user 的 ID 定义单独的类型，防止使用 string 时候容易传错数据。这样在编译期就能检查到错误
type IDUser string
type User struct {
	UserTable
	sq.DefaultLifeCycle
	ID IDUser `db:"id"`
}
// 务必使用 *User，否则不会生效
func (data *User) BeforeCreate() error {
	// 务必判断 ID 为空才赋值，有些业务场景下是手动传递 id 给 Model
	if data.ID == "" {
		data.ID = IDUser(sq.UUID())
	}
	return nil
}
