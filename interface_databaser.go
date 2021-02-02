package sq

import (
	"context"
	"database/sql"
)

type Databaser interface {
	// 配置 SQL 验证器
	SetSQLChecker(sqlChecker SQLChecker)
	// 关闭数据库连接
	Close() error
	// 插入数据
	Insert(ctx context.Context, qb QB) (result sql.Result, err error)
	// 基于 Model 创建数据(insert)
	CreateModel(ctx context.Context, ptr Model, checkSQL ...string) (err error)
	// 查询单行多列
	QueryRowScan(ctx context.Context, qb QB, desc ...interface{}) (has bool, err error)
	// 查询多行多列
	SelectScan(ctx context.Context, qb QB, scan ScanFunc) (err error)
	// 查询单行多列并转换为结构体
	QueryRowStructScan(ctx context.Context, ptr interface{}, qb QB)  (has bool, err error)
	// 查询多行并转换为结构体
	SelectSlice(ctx context.Context, slicePtr interface{}, qb QB) (err error)
	// count
	Count(ctx context.Context, qb QB) (count int, err error)
	// 查询数据是否存在(单条数据是否存在不建议使用 count 而是使用 Exist)
	Exist(ctx context.Context, qb QB) (existed bool, err error)
	// sum
	Sum(ctx context.Context, column Column ,qb QB) (value sql.NullInt64, err error)
	// 查询单条数据并转换为 Model
	QueryModel(ctx context.Context, ptr Model, qb QB) (has bool , err error)
	// 查询多条数据并转换为 Model slice
	QueryModelSlice(ctx context.Context, modelSlicePtr interface{}, qb QB) (err error)
	// 更新
	Update(ctx context.Context, qb QB) (result sql.Result, err error)
	// 基于 Model 更新数据
	UpdateModel(ctx context.Context, ptr Model, updateData []Data, where []Condition, checkSQL ...string) (result sql.Result, err error)
	// 硬删除（不可恢复）
	HardDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (result sql.Result, err error)
	// 软删除（可恢复）
	SoftDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (result sql.Result, err error)
	// 查询单条数据并转换为 Relation
	QueryRelation(ctx context.Context, ptr Relation, qb QB, checkSQL ...string) (has bool, err error)
	// 查询多条数据并转换为 Relation slice
	QueryRelationSlice(ctx context.Context, relationSlicePtr interface{}, qb QB, checkSQL ...string) (err error)
	// 执行SQL
	Exec(ctx context.Context, raw Raw) (result sql.Result, err error)
	// 开启事务
	Transaction(ctx context.Context, handle func (tx *Transaction) TxResult) (err error)
	// 开启自定义级别的事务
	TransactionOpts(ctx context.Context, handle func (tx *Transaction) TxResult, opts *sql.TxOptions) (err error)
}
func verifyDatabaser() {
	db := &Database{}
	func (databaser Databaser) {

	}(db)
}

