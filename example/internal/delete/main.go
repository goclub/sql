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
	// 通过 InsertModel 准备数据
	insertUser := m.User{
		Name:          "delete1",
		Mobile:        "13400001111",
		ChinaIDCardNo: "340828199912121111",
	}
	err = db.InsertModel(ctx, &insertUser, sq.QB{
		UseInsertIgnoreInto: true,
	})
	if err != nil {
		return
	}
	userID := insertUser.ID
	// 软删
	err = db.SoftDelete(ctx, sq.QB{
		From: &m.TableUser{},
		Where: sq.
			And(col.ID, sq.Equal(userID)),
		Review: "TODO",
		Limit:  1,
	})
	if err != nil {
		return
	}
	// 你还可以通过 db.HardDelete() 永久删除数据
	return
}
