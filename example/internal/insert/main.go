package main

import (
	"context"
	sq "github.com/goclub/sql"
	connectMysql "github.com/goclub/sql/example/internal/db"
	m "github.com/goclub/sql/example/internal/model"
	"log"
)

func main () {
	ctx := context.Background()
	err := example(ctx) ; if err != nil {
		log.Print(err)
	}
}
func example(ctx context.Context) (err error) {
	db := connectMysql.DB
	// qb 是 goclub/sql 的核心，用于生产sql
	col := m.TableUser{}.Column()
	qb := sq.QB{
		From: &m.TableUser{},
		Insert: []sq.Insert{
			sq.Value(col.Name, "nimo"),
			sq.Value(col.Mobile, 1341111222),
			sq.Value(col.ChinaIDCardNo, "31111119921219000"),
		},
		// CheckSQL 的作用是用于审查 sql 或增加代码可读性，可以忽略
		Review: "INSERT INTO `user` (`name`,`mobile`,`china_id_card_no`) VALUES (?,?,?)",
	}
	result, err := db.Insert(ctx, qb) ; if err != nil {
		// 无法处理的错误应当向上传递
		return err
	}
	// result.LastInsertId() 一般在 存在 AUTO_INCREMENT 的场景使用
	rowsAffected, err := result.RowsAffected() ; if err != nil {
		// RowsAffected 不是每个数据库或数据库驱动都支持的，需要根据时间情况决定是否要使用 RowsAffected
		return err
	}
	log.Print("rowsAffected:", rowsAffected)
	return nil
}