package main

import (
	"context"
	sq "github.com/goclub/sql"
	connectMysql "github.com/goclub/sql/example/internal/db"
	m "github.com/goclub/sql/example/internal/model"
	"log"
)

func main() {
	ctx := context.Background()
	err := example(ctx)
	if err != nil {
		log.Print(err)
	}
}
func example(ctx context.Context) (err error) {
	db := connectMysql.DB
	// qb 是 goclub/sql 的核心，用于生成sql
	col := m.TableUser{}.Column()
	qb := sq.QB{
		// From 用来配置表名和软删字段
		// From 可以使用 &m.User{} 或者  &m.TableUser{}, 它们两个都是通过 https://goclub.run/?k=model 生成的
		From: &m.TableUser{},
		Insert: sq.Values{
			{col.Name, "nimo"},
			{col.Mobile, "1341111222"},
			{col.ChinaIDCardNo, "31111119921219000"},
		},
		// Review 的作用是用于审查 sql 或增加代码可读性，可以忽略
		Review: "INSERT INTO `user` (`name`,`mobile`,`china_id_card_no`) VALUES (?,?,?)",
	}
	_, err = db.Insert(ctx, qb)
	if err != nil {
		// 无法处理的错误应当向上传递
		return err
	}
	return
}
