package pd

import (
	sq "github.com/goclub/sql"
)
// 配置 Table 可通过继承创建出 Model
type UserTable struct {
	// 当你的表存在 `deleted_at` timestamp NULL DEFAULT NULL 时候 使用sq.SoftDeleteDeletedAt
	// 因为它实现了 SoftDeleteWhere() Raw 方法 和  SoftDeleteSet() Raw 方法
	sq.SoftDeleteDeletedAt
	// 还可以根据表的情况选择 sq.SoftDeleteDeleteTime 或 SoftDeleteIsDeleted

}
// 配置表名
// 使用 (*UserTable) 而不是 (UserTable) 是因为这样可以避免在 InsertModel UpdateModel 等操作的时候忘记传递指针。
func (*UserTable) TableName() string {
	return "user"
}
// 配置字段字典，防止输错字符导致的bug
func (UserTable) Column() (col struct {
	ID sq.Column
	Name sq.Column
	Age sq.Column
	CreatedAt sq.Column
	UpdatedAt sq.Column
	DeletedAt sq.Column
}) {
	col.ID = "id"
	col.Name = "name"
	col.Age = "age"
	col.CreatedAt = "created_at"
	col.UpdatedAt = "updated_at"
	col.DeletedAt = "deleted_at"
	return
}