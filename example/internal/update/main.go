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
	col := m.TableUser{}.Column()
	// qb 是 goclub/sql 的核心，用于生成sql
	qb := sq.QB{
		From: &m.TableUser{},
		// 可以使用 sq.Set 或者  sq.SetRaw
		Update: sq.
			Set(col.Name, "tim").
			Set(col.Mobile, "13022228888"),
		Where: sq.And(col.ID, sq.Equal("1514f086-692e-4666-8bfd-3052d1b51261")),
		// Review 的作用是用于审查 sql 或增加代码可读性，可以忽略
		Review: "UPDATE `user` SET `name`= ?,`mobile`= ? WHERE `id` = ? AND `deleted_at` IS NULL",
	}
	result, err := db.Update(ctx, qb) ; if err != nil {
		// 无法处理的错误应当向上传递
		return
	}
	affected, err := result.RowsAffected() ; if err != nil {
	    return
	}
	log.Print("affected:", affected)
	return
}