package sq_test

import (
	"context"
	"database/sql"
	sq "github.com/goclub/sql"
	"log"
)

func ExampleSqlx() {
	ExampleSqlx_QueryxRowScanStruct()
	ExampleSqlx_QueryRowxCount()
	ExampleSqlx_QueryRowxScan()
	ExampleSqlx_QueryxScan()

	ExampleSqlx_Select()
	ExampleSqlx_ExecUpdate()
	ExampleSqlx_Insert()
}

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
	log.Print("ExampleDB_Select")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	var nameAgeList []struct {
		Name string `db:"name"`
		Age int `db:"age"`
	}

	selectSQL := "SELECT `name`,`age` FROM `user`"
	err := exampleDB.Core.SelectContext(ctx, &nameAgeList, selectSQL) ; if err != nil {
		panic(err)
	}
	log.Print(nameAgeList)
}
func ExampleSqlx_ExecUpdate() {
	ctx := context.TODO()
	updateSQL := "UPDATE `user` SET `name` = ?, `age` = ? WHERE `name` = ?"
	_, err := exampleDB.Core.ExecContext(ctx, updateSQL, "newName", 20, "oldName") ; if err != nil {
		panic(err)
	}
}
func ExampleSqlx_Insert() {
	ctx := context.TODO()
	insertSQL := "INSERT INTO `user` (`id`, `name`, `age`) VALUES(?, ?, ?)"
	values := []interface{}{sq.UUID(), "newName", 18}
	_, err := exampleDB.Core.ExecContext(ctx, insertSQL, values...) ; if err != nil {
		panic(err)
	}
}
