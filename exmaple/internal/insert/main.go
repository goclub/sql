package main

import (
	"context"
	sq "github.com/goclub/sql"
	connectMysql "github.com/goclub/sql/exmaple/internal/db"
	"github.com/goclub/sql/exmaple/internal/pd"

	"log"
)

func main () {
	ctx := context.Background()
	err := example(ctx) ; if err != nil {
		log.Print(err)
	}
}
func example(ctx context.Context) error {
	checkSQL := []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"}
	db := connectMysql.DB
	// qb 是 goclub/sql 的核心，用于生产sql
	userCol := pd.UserTable{}.Column()
	qb := sq.QB{
		Table: &pd.UserTable{},
		Insert: []sq.Insert{
			sq.Value(userCol.ID, sq.UUID()),
			sq.Value(userCol.Name, "nimo"),
			sq.Value(userCol.Age, 18),
		},
		// CheckSQL 的作用是给 DBA 审查 sql 或增加代码可读性，可以不传
		CheckSQL: checkSQL,
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