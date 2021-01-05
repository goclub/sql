package sq_test

import (
	sq "github.com/goclub/sql"
)


type TableUser struct {
	sq.SoftDeleteDeletedAt
}
func (TableUser) TableName() string {return "user"}
type IDUser string
type User struct {
	TableUser
	sq.DefaultModel
	ID IDUser `db:"id"`
	Name string `db:"name"`
	Age int `db:"age"`
}
func (u *User) BeforeCreate() error {
	if len(u.ID) == 0 {
		u.ID = IDUser(sq.UUID())
	}
	return nil
}
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
