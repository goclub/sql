// Generate by https://goclub.run
package m

import (
	sq "github.com/goclub/sql"
)

type TableUserAddress struct {
	sq.WithoutSoftDelete
}
// 给 TableName 加上指针 * 能避免 db.InsertModel(user) 这种错误， 应当使用 db.InsertModel(&user) 或
func (*TableUserAddress) TableName() string { return "user_address" }
type UserAddress struct {
	UserID  IDUser  `db:"user_id"`
	Address string  `db:"address"`
	TableUserAddress
	sq.CreatedAtUpdatedAt
	sq.DefaultLifeCycle
}
func (v UserAddress) PrimaryKey() []sq.Condition {
	return sq.And(
		v.Column().UserID, sq.Equal(v.UserID),
	)
}

func (v TableUserAddress) Column() (col struct{
	UserID   sq.Column
	Address  sq.Column
	CreatedAt sq.Column
	UpdatedAt sq.Column
}) {
	col.UserID   = "user_id"
	col.Address  = "address"
	col.CreatedAt = "created_at"
	col.UpdatedAt = "updated_at"
	return
}
