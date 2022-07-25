package sq

import (
	"context"
	"database/sql"
)

type API interface {
	// 检查连通性
	Ping(ctx context.Context) error
	// 配置 SQL 验证器
	SetSQLChecker(sqlChecker SQLChecker)
	// 关闭数据库连接
	Close() error

	// 插入数据
	Insert(ctx context.Context, qb QB) (result Result, err error)
	// 基于 Model 创建数据, 根据 Model 字段自动填充 qb.Insert
	InsertModel(ctx context.Context, ptr Model, qb QB) (result Result, err error)
	// 查询单行多列 类似 sql.Row{}.Scan()
	QueryRowScan(ctx context.Context, qb QB, desc []interface{}) (has bool, err error)
	// 查询单行多列并转换为结构体
	Query(ctx context.Context, ptr Tabler, qb QB) (has bool, err error)
	// 查询多行并转换为结构体
	QuerySlice(ctx context.Context, slicePtr interface{}, qb QB) (err error)
	// 查询多行多列(自定义扫描)
	QuerySliceScaner(ctx context.Context, qb QB, scaner Scaner) (err error)

	QueryRelation(ctx context.Context, ptr Relation, qb QB) (has bool, err error)
	// 查询多条数据并转换为 Relation slice
	QueryRelationSlice(ctx context.Context, relationSlicePtr interface{}, qb QB) (err error)

	// Count
	Count(ctx context.Context, from Tabler, qb QB) (count uint64, err error)
	// 查询数据是否存在(单条数据是否存在不建议使用 count 而是使用 Exist)
	Has(ctx context.Context, from Tabler, qb QB) (has bool, err error)
	SumInt64(ctx context.Context, from Tabler, column Column, qb QB) (value sql.NullInt64, err error)
	SumFloat64(ctx context.Context, from Tabler, column Column, qb QB) (value sql.NullFloat64, err error)
	// 查询单条数据并转换为 Model

	// 更新
	Update(ctx context.Context, qb QB) (result Result, err error)
	// 暂时取消 QueryModel 因为容易产生歧义
	// // 基于 Model 的 PrimaryKey 更新数据 并自动处理 create update 字段
	// UpdateModel(ctx context.Context, ptr Model, updateData []Set, qb QB) (result Result, err error)

	// 删除测试数据库的数据，只能运行在 test_ 为前缀的数据库中
	ClearTestData(ctx context.Context, qb QB) (result Result, err error)
	// // 基于 Model 删除测试数据库的数据，只能运行在 test_ 为前缀的数据库中
	// ClearTestModel(ctx context.Context, model Model, qb QB) (result Result, err error)

	// 硬删除（不可恢复）
	HardDelete(ctx context.Context, qb QB) (result Result, err error)

	// 软删除（可恢复）
	SoftDelete(ctx context.Context, qb QB) (result Result, err error)

	// 执行QB
	ExecQB(ctx context.Context, qb QB, statement Statement) (result Result, err error)
	// 执行
	Exec(ctx context.Context, query string, values []interface{}) (result Result, err error)

	// 开启事务
	BeginTransaction(ctx context.Context, level sql.IsolationLevel, handle func(tx *Transaction) TxResult) (rollbackNoError bool, err error)
	BeginTransactionOpt(ctx context.Context, opt sql.TxOptions, handle func(tx *Transaction) TxResult) (rollbackNoError bool, err error)
	// show status like "last_query_cost"
	LastQueryCost(ctx context.Context) (lastQueryCost float64, err error)
	// 打印 show status like "last_query_cost" 的结果
	PrintLastQueryCost(ctx context.Context)
}

func verifyDoc() {
	db := &Database{}
	func(API) {

	}(db)
	tx := struct {
		Transaction
		onlyDB
	}{}
	func(API) {

	}(&tx)
}

type onlyDB struct{}

func (onlyDB) Ping(ctx context.Context) error {
	return nil
}
func (onlyDB) SetSQLChecker(sqlChecker SQLChecker) {

}
func (onlyDB) Close() error {
	return nil
}
func (onlyDB) ClearTestData(ctx context.Context, qb QB) (result Result, err error) {
	return
}

// func (onlyDB) ClearTestModel(ctx context.Context, model Model, qb QB) (result Result, err error) {
// 	return
// }
func (onlyDB) BeginTransaction(ctx context.Context, level sql.IsolationLevel, handle func(tx *Transaction) TxResult) (rollbackNoError bool, err error) {
	return
}
func (onlyDB) BeginTransactionOpt(ctx context.Context, opt sql.TxOptions, handle func(tx *Transaction) TxResult) (rollbackNoError bool, err error) {
	return
}
