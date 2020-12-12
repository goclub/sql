package sq_test

import (
	"context"
	"database/sql"
	"github.com/goclub/sql"
	"log"
)
func ExampleOpenAndClose() {
	db, dbClose, err := sq.Open("mysql", sq.DataSourceName{
		DriverName: "mysql",
		User: "root",
		Password:"password",
		Host: "127.0.0.1",
		Port:"3306",
		DB: "goclub_sq",
	}.String()) ; if err != nil {
		panic(err)
	}
	defer dbClose() // 不要频繁的 Open Close
	err = db.Core.Ping() ; if err != nil {
		panic(err)
	}
}

type User struct {
	ID string `db:"id"`
	Name string `db:"name"`
	Age string `db:"age"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}
func (User) TableName() string {return "user"}
func (u *User) BeforeCreate() {
	if len(u.ID) == 0 { u.ID = sq.UUID() }
}
func ExampleQB() {
	db := sq.DB{}
	ctx := context.Background()
	// 查询一条数据
	{
		user := User{}
		hasUser, err := db.One(ctx, &user, sq.QB{
			Where: sq.
				And("name", sq.Equal("nimo")).
				And("age", sq.Equal(18)),
			Check:[]string{
				"SELECT `id`,`name`,`age`,`deleted_at` FROM `user` WHERE `name` = ? AND `age` = ? AND deleted_at IS NULL limit 1",
			},
		}) ; if err != nil {
		panic(err)
		}
		log.Print(user, hasUser)
	}
	{
		var count int
		err := db.Scan(ctx, sq.QB{
			Table: User{}.TableName(),
			Select: []sq.Column{"COUNT(*)"},
			Where: sq.And("age", sq.GtInt(18)),
		}, &count) ; if err != nil {
		panic(err)
	}
		log.Print(count)
	}
	{
		userNameAge := struct {
			Name string `db:"name"`
			Age int `db:"age"`
		}{}
		hasUser, err := db.ScanStruct(ctx, &userNameAge, sq.QB{
			Table: User{}.TableName(),
			Where: sq.
				And("name", sq.Equal("nimo")).
				And("age", sq.Equal(18)),
			Check:[]string{
				"SELECT `name`,`age`,`deleted_at` FROM `user` WHERE `name` = ? AND `age` = ? AND deleted_at IS NULL limit 1",
			},
		}) ; if err != nil {
		panic(err)
	}
		log.Print(userNameAge, hasUser)
	}
	// 查询多条数据
	{
		var userList []User
		err := db.List(ctx, &userList, sq.QB{
			Where: sq.
				And("age", sq.GtInt(18)),
			Check:[]string{
				"SELECT `id`,`name`,`age`,`deleted_at` FROM `user` WHERE `name` = ? AND `age` = ? AND deleted_at IS NULL",
			},
		}) ; if err != nil {
			panic(err)
		}
		log.Print(userList)
	}
	{
		var userList []struct {
			Name string `db:"name"`
			Age int `db:"age"`
		}
		err := db.Select(ctx, &userList, sq.QB{
			Table: User{}.TableName(),
			Where: sq.
				And("age", sq.GtInt(18)),
			Check:[]string{
				"SELECT `name`,`age` FROM `user` WHERE `age` = ? AND deleted_at IS NULL",
			},
		}) ; if err != nil {
		panic(err)
	}
		log.Print(userList)
	}
	{
		count, err := db.Count(ctx, &User{}, sq.QB{
			Where: sq.And("age", sq.GtInt(18)),
		}) ; if err != nil {
			panic(err)
		}
		log.Print(count)
	}
}