package sq_test

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goclub/sql"
	"log"
	"testing"
)

func TestExample(t *testing.T) {
	ExampleCoreQueryRowx()
	ExampleCoreQueryxScan()
	ExampleCoreQueryxScanStruct()
	ExampleCoreQueryRowxCount()
}
var exampleDB *sq.DB
func init () {
	db, dbClose, err := sq.Open("mysql", sq.DataSourceName{
		DriverName: "mysql",
		User: "root",
		Password:"somepass",
		Host: "127.0.0.1",
		Port:"3306",
		DB: "goclub_sql",
	}.String()) ; if err != nil {
		panic(err)
	}
	exampleDB = db
	_=dbClose // init 场景下不需要 close，应该在 main 执行完毕后 close
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

func ExampleCoreQueryRowx() {
	log.Print("ExampleCoreQueryRowx")
	var name string
	var has bool
	// 虽然 row 只 scan 一条数据，但是还是加上 LIMIT 1 以确保最高性能
	row := exampleDB.Core.QueryRowx(`SELECT name FROM query WHERE id = ? LIMIT 1`, 1)
	err := row.Scan(&name) ; if err != nil {
		if err == sql.ErrNoRows {
			has = false
		} else {
			panic(err) // 项目中应该 return err 将可处理的错误传递
		}
	} else {
		has = true
	}
	log.Print(name, has)
}
func ExampleCoreQueryRowxCount() {
	log.Print("ExampleCoreQueryRowxCount")
	var count int
	row := exampleDB.Core.QueryRowx(`SELECT COUNT(*) FROM query`)
	err := row.Scan(&count) ; if err != nil {
		if err == sql.ErrNoRows {
			panic(err) // 虽然 count 必然会有结果，但是还是做个判断（防御措施）
		} else {
			panic(err) // 项目中应该 return err 将可处理的错误传递
		}
	}
	log.Print(count)
}
func ExampleCoreQueryxScan() {
	log.Print("ExampleCoreQueryx")
	rows, err := exampleDB.Core.Queryx(`SELECT id,name FROM query WHERE name like ?`, `%m%`) ; if err != nil {
		panic(err)
	}
	if rows != nil {
		defer rows.Close()
	}
	type data struct {
		ID string
		Name string
	}
	var list []data
	for rows.Next() {
		var data data
		err := rows.Scan(&data.ID, &data.Name) ; if err != nil {
			panic(err) // 项目中应该 return err 将可处理的错误传递
		}
		list = append(list, data)
	}
	err = rows.Err() ; if err != nil {
		panic(err) // 项目中应该 return err 将可处理的错误传递
	}
	log.Print(list)
}

func ExampleCoreQueryxScanStruct() {
	log.Print("ExampleCoreQueryxScanStruct")
	rows, err := exampleDB.Core.Queryx(`SELECT id,name FROM query WHERE name like ?`, `%m%`) ; if err != nil {
		panic(err)
	}
	if rows != nil {
		defer rows.Close()
	}
	type data struct {
		ID string `db:"id"`
		Name string `db:"name"`
	}
	var list []data
	for rows.Next() {
		var data data
		// 使用 StructScan 时候 会基于结构体中的 db 结构体标签作为 scan 的标识
		err := rows.StructScan(&data) ; if err != nil {
			panic(err) // 项目中应该 return err 将可处理的错误传递
		}
		list = append(list, data)
	}
	err = rows.Err() ; if err != nil {
		panic(err) // 项目中应该 return err 将可处理的错误传递
	}
	log.Print(list)
}