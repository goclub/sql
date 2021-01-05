package sq_test

import (
	"database/sql"
	sq "github.com/goclub/sql"
	"time"
)
type UserTable struct {
	sq.SoftDeleteDeletedAt
}
func (UserTable) TableName() string {return "user"}
type IDUser string
type User struct {
	ID IDUser `db:"id"`
	Name string `db:"name"`
	Age int `db:"age"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at" sq:"ignore"`
	UserTable
}

func (u *User) BeforeCreate() {if len(u.ID) == 0 { u.ID = IDUser(sq.UUID()) }}
func (u *User) AfterCreate(result sql.Result) (err error) {return nil}
func (u *User) BeforeUpdate(){}
func (u *User) AfterUpdate(){}
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
