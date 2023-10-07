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
	user := m.User{
		Name:          "nimo",
		Mobile:        "1341111222",
		ChinaIDCardNo: "31111119921219000",
	}
	// InsertModel 会自动获取 user 的结构体字段(struct field),会忽略 struct tag 中带有 sq:"ignoreInsert" 的字段
	// 使用 InsertModel 时 sq.QB 不需要配置 Form
	err = db.InsertModel(ctx, &user, sq.QB{
		// Review 的作用是用于审查 sql 或增加代码可读性，可以忽略
		Review: "INSERT INTO `user` (`name`,`mobile`,`china_id_card_no`,`created_at`,`updated_at`) VALUES (?,?,?,?,?)",
	})
	if err != nil {
		return
	}
	return nil
}
