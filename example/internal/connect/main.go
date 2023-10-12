package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	sq "github.com/goclub/sql"
	"log"
	"time"
)

var db *sq.Database

func init() {
	var err error
	var dbClose func() error
	db, dbClose, err = sq.Open("mysql", sq.MysqlDataSource{
		// 生产环境请使用环境变量或者配置中心配置数据库地址，不要硬编码在代码中
		User:     "root",
		Password: "somepass",
		Host:     "127.0.0.1",
		Port:     "3306",
		DB:       "example_goclub_sql",
		Query: map[string]string{
			"charset":   "utf8",
			"parseTime": "True",
			"loc":       "Local",
		},
	}.FormatDSN())
	if err != nil {
		// 大部分创建数据库连接失败应该panic
		panic(err)
	}
	// 使用 init 方式连接数据库则无需 close ，程序退出再执行close
	_ = dbClose()
}
func main() {
	ctx := context.Background()
	// 设置ping超时1s则视为失败
	pingCtx, cancelFunc := context.WithTimeout(ctx, time.Second)
	defer cancelFunc()
	err := db.Ping(pingCtx)
	if err != nil {
		panic(err)
	}
	log.Print("连接成功")
}
