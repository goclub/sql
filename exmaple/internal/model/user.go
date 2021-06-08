
// Generate by https://tools.goclub.vip
package m
import (
	"database/sql"
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
	ID   IDUser  `db:"id" sq:"ignoreCreate" sq:"ignoreUpdate" `
	Name string  `db:"name"`
	Age  uint8   `db:"age"`
	TableUser
	sq.CreatedAtUpdatedAt
	sq.DefaultLifeCycle
}
func (v User) PrimaryKey() []sq.Condition {
	return sq.And(
		v.Column().ID, sq.Equal(v.ID),
	)
}

func (v *User) AfterCreate(result sql.Result) error {
	id, err := result.LastInsertId(); if err != nil {
		return err
	}
	v.ID = IDUser(uint64(id))
	return nil
}

func (v TableUser) Column() (col struct{
	ID    sq.Column
	Name  sq.Column
	Age   sq.Column
	CreatedAt sq.Column
	UpdatedAt sq.Column
}) {
	col.ID    = "id"
	col.Name  = "name"
	col.Age   = "age"
	col.CreatedAt = "created_at"
	col.UpdatedAt = "updated_at"
	return
}
