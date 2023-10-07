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
	col := m.TableUser{}.Column()
	// qb 是 goclub/sql 的核心，用于生成sql
	qb := sq.QB{
		From: &m.TableUser{},
		// 可以使用 sq.SetMap/sq.Set/sq.SetRaw
		Set: sq.SetMap(map[sq.Column]interface{}{
			col.Name:   "tim",
			col.Mobile: "13022228888",
		}),
		Where: sq.And(col.ID, sq.Equal("1514f086-692e-4666-8bfd-3052d1b51261")),
		// Review 的作用是用于审查 sql 或增加代码可读性，可以忽略
		Review: "UPDATE `user` SET `mobile`= ?,`name`= ? WHERE `id` = ? AND `deleted_at` IS NULL",
	}
	affected, err := db.UpdateAffected(ctx, &m.TableUser{}, qb)
	if err != nil {
		// 无法处理的错误应当向上传递
		return
	}
	if err != nil {
		return
	}
	log.Print("affected:", affected)

	// 你可以直接简写成
	// affected, err := sq.RowsAffected(db.Set(ctx, qb)) ; if err != nil {
	// 	return
	// }
	return
}
