package connectMysql

import (
	_ "github.com/go-sql-driver/mysql"
	sq "github.com/goclub/sql"
)

var DB *sq.Database
func init () {
	var err error
	var dbClose func() error
	DB, dbClose, err = sq.Open("mysql", sq.MysqlDataSource{
		// 生产环境请使用环境变量或者配置中心配置数据库地址，不要硬编码在代码中
		User:     "root",
		Password: "somepass",
		Host:     "127.0.0.1",
		Port:     "3306",
		DB:       "example_goclub_sql",
		Query: map[string]string{
			"charset": "utf8",
			"parseTime": "True",
			"loc": "Local",
		},
	}.String()) ; if err != nil {
		// 大部分创建数据库连接失败应该panic
		panic(err)
	}
	// 使用 init 方式连接数据库则无需 close ，依赖注入场景下才需要 close
	_ = dbClose()
}