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
	ExampleDB_QueryRowScan()
	ExampleDB_QueryRowScanMultiColumn()
	ExampleDB_QueryRowStructScan()
	ExampleDB_One()
	ExampleDB_Count()

	ExampleSqlx_QueryxRowScanStruct()
	ExampleSqlx_QueryRowxCount()
	ExampleSqlx_QueryRowxScan()
	ExampleSqlx_QueryxScan()

	ExampleSqlx_Select()
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
	err = exampleDB.Core.Ping() ; if err != nil {
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
func (User) Column () (col struct{
	ID sq.Column
	Name sq.Column
	Age sq.Column
	DeletedAt sq.Column
}) {
	col.ID = "id"
	col.Name = "name"
	col.Age = "age"
	col.DeletedAt = "deleted_at"
	return
}
// 查询单行单列数据
func ExampleDB_QueryRowScan() {
	log.Print("ExampleDB_QueryRowScan")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userCol := User{}.Column()
	var name string
	qb := sq.QB{
		Table: User{}.TableName(),
		Select: []sq.Column{userCol.Name},
		Where: sq.
			And(userCol.ID, sq.Equal(1)),
	}.Check("SELECT `name` FROM `user` WHERE `id` = ? LIMIT ?")
	hasName, err := exampleDB.QueryRowScan(ctx, qb, &name) ; if err != nil {
		panic(err)
	}
	log.Print(name, hasName)
}
// 查询单行多列数据
func ExampleDB_QueryRowScanMultiColumn() {
	log.Print("ExampleDB_QueryRowScanMultiColumn")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userCol := User{}.Column()
	var name string
	var age int
	qb := sq.QB{
		Table: User{}.TableName(),
		Select: []sq.Column{userCol.Name, userCol.Age},
		Where: sq.
			And(userCol.ID, sq.Equal(1)),
	}.Check("SELECT `name`,`age` FROM `user` WHERE `id` = ? LIMIT ?")
	hasName, err := exampleDB.QueryRowScan(ctx, qb, &name,&age) ; if err != nil {
		panic(err)
	}
	log.Print(name, hasName)
}
// 查询单行多列数据(扫描到结构体)
func ExampleDB_QueryRowStructScan() {
	log.Print("ExampleDB_QueryRowStructScan")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userNameAge := struct {
		Name string `db:"name"`
		Age int `db:"age"`
	}{}
	userCol := User{}.Column()
	qb := sq.QB{
		Table: User{}.TableName(),
		Where: sq.
			And(userCol.Name, sq.Equal("nimo")).
			And(userCol.Age, sq.Equal(18)),
	}.Check("SELECT `name`,`age` FROM `user` WHERE `name` = ? AND `age` = ? AND deleted_at IS NULL LIMIT ?")
	hasUser, err := exampleDB.QueryRowStructScan(ctx, &userNameAge, qb) ; if err != nil {
		panic(err)
	}
	log.Print(userNameAge, hasUser)
}

// 基于 Model 查询单行数据 （可省略 qb.Table）
func ExampleDB_One() {
	log.Print("ExampleDB_One")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	user := User{}
	userCol := user.Column()
	qb := sq.QB{
		Where: sq.
			And(userCol.Name, sq.Equal("nimo")).
			And(userCol.Age, sq.Equal(18)),
	}.Check("SELECT `id`,`name`,`age`,`deleted_at` FROM `user` WHERE `name` = ? AND `age` = ? AND deleted_at IS NULL LIMIT ?")
	hasUser, err := exampleDB.One(ctx, &user, qb) ; if err != nil {
		panic(err)
	}
	log.Print(user, hasUser)
}
// select count(*) from table 便捷方法
func ExampleDB_Count() {
	log.Print("ExampleDB_Count")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	count, err := exampleDB.Count(ctx, &User{}, sq.QB{
		Where: sq.And(User{}.Column().Age, sq.GtInt(18)),
	}) ; if err != nil {
		panic(err)
	}
	log.Print(count)
}

// // query builder 的示例
// func ExampleQB() {
// 	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
// 	// 查询多条数据
// 	{
// 		var userList []User
// 		err := exampleDB.List(ctx, &userList, sq.QB{
// 			Where: sq.
// 				And("age", sq.GtInt(18)),
// 		}.Check("SELECT `id`,`name`,`age`,`deleted_at` FROM `user` WHERE `name` = ? AND `age` = ? AND deleted_at IS NULL")) ; if err != nil {
// 			panic(err)
// 		}
// 		log.Print(userList)
// 	}
// 	{
// 		var userList []struct {
// 			Name string `db:"name"`
// 			Age int `db:"age"`
// 		}
// 		err := exampleDB.Select(ctx, &userList, sq.QB{
// 			Table: User{}.TableName(),
// 			Where: sq.
// 				And("age", sq.GtInt(18)),
// 		}.Check("SELECT `name`,`age` FROM `user` WHERE `age` = ? AND deleted_at IS NULL")) ; if err != nil {
// 		panic(err)
// 	}
// 		log.Print(userList)
// 	}
// 	{
// 		count, err := exampleDB.Count(ctx, &User{}, sq.QB{
// 			Where: sq.And("age", sq.GtInt(18)),
// 		}) ; if err != nil {
// 			panic(err)
// 		}
// 		log.Print(count)
// 	}
// }

// sqlx 的单行数据查询
func ExampleSqlx_QueryRowxScan() {
	log.Print("ExampleSqlx_QueryRowxScan")
	var name string
	var has bool
	// 虽然 row 只 scan 一条数据，但是还是加上 LIMIT ? 以确保最高性能
	row := exampleDB.Core.QueryRowx(`SELECT name FROM query WHERE id = ? LIMIT ?`, 1, 1)
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
// sqlx 的count 查询
func ExampleSqlx_QueryRowxCount() {
	log.Print("ExampleSqlx_QueryRowxCount")
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
// sqlx 的多行数据查询（扫描结构体）
func ExampleSqlx_QueryxScan() {
	log.Print("ExampleSqlx_QueryxRowScan")
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

// sqlx的多行数据查询（扫描结构体）
func ExampleSqlx_QueryxRowScanStruct() {
	log.Print("ExampleSqlx_QueryxRowScanStruct")
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
func ExampleSqlx_Select() {
	log.Print("ExampleSqlx_Select")
	var userList []User 
	err := exampleDB.Core.Select(&userList, "SELECT `id`, `name` FROM user WHERE `name` LIKE ?", `%m%`) ; if err != nil {
		panic(err)
	}
	log.Print(userList)
}