package main

import (
	_ "github.com/go-sql-driver/mysql"
	sq "github.com/goclub/sql"
	migrateActions "github.com/goclub/sql/exmaple/migrate/actions"
)

func main() {
	db, dbClose, err := sq.Open("mysql", sq.DataSourceName{
		DriverName: "mysql",
		User: "root",
		Password:"somepass",
		Host: "127.0.0.1",
		Port:"3306",
		DB: "test_goclub_sql",
	}.String()) ; if err != nil {
		panic(err)
	}
	defer dbClose()
	sq.ExecMigrate(db, &migrateActions.Migrate{})
}
