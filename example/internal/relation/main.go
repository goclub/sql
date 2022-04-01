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
	// 准备数据
	var insertUser m.User
	// 多表插入一定要用事务,否则无法保证数据一致性
	rollbackNoError, err := db.BeginTransaction(ctx, sq.LevelReadCommitted, func(tx *sq.Transaction) sq.TxResult {
		db := false ; _=db // 一般情况下事务中都是使用tx所以重新声明变量db 防止在 tx 中使用db
		insertUser =  m.User{
			Name: "relation mobile",
			Mobile: "13411122222",
			ChinaIDCardNo: "310113199912121112",
		}
		_, err = tx.InsertModel(ctx, &insertUser, sq.QB{
			UseInsertIgnoreInto: true, // 为了便于测试忽略重复插入
		}) ; if err != nil {
			return tx.RollbackWithError(err)
		}
		_, err = tx.InsertModel(ctx, &m.UserAddress{
			UserID:             insertUser.ID,
			Address:            "天堂路",
		}, sq.QB{
			UseInsertIgnoreInto: true, // 为了便于测试忽略重复插入
		}) ; if err != nil {
			return tx.RollbackWithError(err)
		}
		return tx.Commit()
	}) ; if err != nil {
	    return
	}
	if rollbackNoError {
		// 运行到 BeginTransaction 中的 return tx.Rollback() 时, rollbackNoError 为 true
	}
	userWithAddress := m.UserWithAddress{}
	col := userWithAddress.Column()
	hasUserWithAddress, err := db.QueryRelation(ctx, &userWithAddress, sq.QB{
		Where: sq.And(col.UserID, sq.Equal(insertUser.ID)),
	}) ; if err != nil {
	    return
	}
	log.Print("hasUserWithAddress:", hasUserWithAddress)
	log.Print("userWithAddress:", userWithAddress)
	// QueryRelationSlice 查询多条关联数据
	var userWithAddressList []m.UserWithAddress
	err = db.QueryRelationSlice(ctx, &userWithAddressList, sq.QB{}) ; if err != nil {
	    return
	}
	log.Print("userWithAddressList:", userWithAddressList)
	return
}