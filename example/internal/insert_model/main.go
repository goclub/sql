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
	user := m.User{
		Name:  "nimo",
		Mobile: "1341111222",
		ChinaIDCardNo: "31111119921219000",
	}
	// InsertModel 会自动获取 user struct field,会忽略 struct tag 中带有 sq:"ignoreUpdate" 的字段
	result, err := db.InsertModel(ctx, &user, sq.QB{
		// Review 可不填
		Review: "INSERT INTO `user` (`name`,`mobile`,`china_id_card_no`,`created_at`,`updated_at`) VALUES (?,?,?,?,?)",
	}) ; if err != nil {
	    return
	}
	// result.LastInsertId() 一般在 存在 AUTO_INCREMENT 的场景使用
	rowsAffected, err := result.RowsAffected() ; if err != nil {
		// RowsAffected 不是每个数据库或数据库驱动都支持的，需要根据时间情况决定是否要使用 RowsAffected
		return err
	}
	log.Print("rowsAffected:", rowsAffected)
	return nil
}