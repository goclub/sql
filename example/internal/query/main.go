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
	// 通过 InsertModel 准备数据
	insertUser :=  m.User{
		Name: "query mobile",
		Mobile: "13411122222",
		ChinaIDCardNo: "310113199912121111",
	}
	_, err = db.InsertModel(ctx, &insertUser, sq.QB{
		UseInsertIgnoreInto: true, // 为了便于测试忽略重复插入
	}) ; if err != nil {
	    return
	}
	userID := insertUser.ID
	// 基于 Model 查询
	user := m.User{}
	// Query 会自动分析 user 的结构体字段(struct field) 这样 sq.QB{}.Select 就可以省略了
	// sq.QB{}.Form 也会自动设置为 user
	hasUser, err := db.Query(ctx, &user, sq.QB{
		Where: sq.
			And(col.ID, sq.Equal(userID)),
		Review: "SELECT `id`, `name`, `mobile`, `china_id_card_no`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?",
	}) ; if err != nil {
	    return
	}
	log.Print("hasUser:", hasUser)
	log.Print("user:", user)
	// 基于 TableUser 查询部分数据
	type PartUser struct {
		m.TableUser // 组合 TableUser 可以快速配置表名和软删
		Name string `db:"name"`
		Mobile string `db:"mobile"`
	}
	partUser := PartUser{}
	hasPartUser, err := db.Query(ctx, &partUser, sq.QB{
		Where: sq.
			And(col.ID, sq.Equal(userID)),
		Review: "SELECT `name`, `mobile` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?",
	}) ; if err != nil {
	    return
	}
	log.Print("hasPartUser:", hasPartUser)
	log.Print("partUser:", partUser)

	// QuerySlice 可以查询多条数据
	var userList []m.User
	err = db.QuerySlice(ctx, &userList, sq.QB{}) ; if err != nil {
	    return
	}
	log.Print("userList:", userList)
	// 还可以使用 sq.QB{}.Paging(1, 10) 进行分页查询
	/*
		err = db.QuerySlice(ctx, &userList, sq.QB{}.Paging(1, 10)) ; if err != nil {
		    return
		}
	*/
	return
}