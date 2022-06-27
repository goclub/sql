// Generate by https://goclub.run
package m

import (
	sq "github.com/goclub/sql"
)

type IDUser uint64
func NewIDUser(id uint64) IDUser {
	return IDUser(id)
}
func (id IDUser) Uint64() uint64 {
	return uint64(id)
}
type TableUser struct {
	sq.SoftDeletedAt
}
// 给 TableName 加上指针 * 能避免 db.InsertModel(user) 这种错误， 应当使用 db.InsertModel(&user) 或
func (*TableUser) TableName() string { return "user" }
type User struct {
	ID            IDUser  `db:"id" sq:"ignoreInsert" sq:"ignoreUpdate" `
	Name          string  `db:"name"`
	Mobile        string  `db:"mobile"`
	ChinaIDCardNo string  `db:"china_id_card_no"`
	TableUser
	sq.CreatedAtUpdatedAt
	sq.DefaultLifeCycle
}


func (v *User) AfterInsert(result sq.Result) error {
	id, err := result.LastInsertUint64Id(); if err != nil {
		return err
	}
	v.ID = NewIDUser(id)
	return nil
}

func (v TableUser) Column() (col struct{
	ID             sq.Column
	Name           sq.Column
	Mobile         sq.Column
	ChinaIDCardNo  sq.Column
	CreatedAt sq.Column
	UpdatedAt sq.Column
}) {
	col.ID             = "id"
	col.Name           = "name"
	col.Mobile         = "mobile"
	col.ChinaIDCardNo  = "china_id_card_no"
	col.CreatedAt = "created_at"
	col.UpdatedAt = "updated_at"
	return
}
