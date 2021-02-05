package sq

import (
	"context"
	"database/sql"
)
type APIDatabase interface {
	// 配置 SQL 验证器
	SetSQLChecker(sqlChecker SQLChecker)
	// 关闭数据库连接
	Close() error

	// 插入数据
	Insert(ctx context.Context, qb QB) (result sql.Result, err error)
	// 基于 Model 创建数据
	ModelInsert(ctx context.Context, ptr Model, checkSQL ...string) (err error)

	// 查询单行多列
	QueryRowScan(ctx context.Context, qb QB, desc ...interface{}) (has bool, err error)
	// 查询单行多列并转换为结构体
	QueryRowStructScan(ctx context.Context, ptr interface{}, qb QB)  (has bool, err error)
	// 查询多行多列
	QueryScan(ctx context.Context, qb QB, scan ScanFunc) (err error)
	// 查询多行并转换为结构体
	QuerySlice(ctx context.Context, slicePtr interface{}, qb QB) (err error)

	// 查询单条数据并转换为 Model slice

	(ctx context.Context, ptr Model, qb QB) (has bool , err error)
	// 查询多条数据并转换为 Model slice
	ModelQuerySlice(ctx context.Context, modelSlicePtr interface{}, qb QB) (err error)
	// 查询单条数据并转换为 Relation
	RelationQueryRow(ctx context.Context, ptr Relation, qb QB) (has bool, err error)
	// 查询多条数据并转换为 Relation slice
	RelationQuerySlice(ctx context.Context, relationSlicePtr interface{}, qb QB) (err error)

	// count
	Count(ctx context.Context, qb QB) (count uint64, err error)
	// 查询数据是否存在(单条数据是否存在不建议使用 count 而是使用 Exist)
	Has(ctx context.Context, qb QB) (has bool, err error)
	// sum
	Sum(ctx context.Context, column Column ,qb QB) (value sql.NullInt64, err error)
	// 查询单条数据并转换为 Model

	// 更新
	Update(ctx context.Context, qb QB) (result sql.Result, err error)
	// 基于 Model 更新数据
	ModelUpdate(ctx context.Context, ptr Model, updateData []Update, where []Condition, checkSQL ...string) (result sql.Result, err error)

	// 硬删除（不可恢复）
	HardDelete(ctx context.Context, qb QB) (result sql.Result, err error)
	// 基于 Model 硬删除（不可恢复）
	ModelHardDelete(ctx context.Context, ptr Model, checkSQL ...string) (result sql.Result, err error)
	// 软删除（可恢复）
	SoftDelete(ctx context.Context, qb QB) (result sql.Result, err error)
	// 基于 Model 软删除（可恢复）
	ModelSoftDelete(ctx context.Context, ptr Model, checkSQL ...string) (result sql.Result, err error)

	// 执行QB
	ExecQB(ctx context.Context, qb QB, statement Statement) (result sql.Result, err error)
	// 执行
	Exec(ctx context.Context, query string, values []interface{}) (result sql.Result, err error)

	// 开启事务
	Transaction(ctx context.Context, handle func (tx *Transaction) TxResult) (isRollback bool, err error)
	// 开启自定义级别的事务
	TransactionOpts(ctx context.Context, handle func (tx *Transaction) TxResult, opts *sql.TxOptions) (isRollback bool, err error)
}
func verifyDoc() {
	db := &Database{}
	func (APIDatabase) {

	}(db)
}

