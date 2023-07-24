package sq

import (
	"context"
	"database/sql"
	xerr "github.com/goclub/error"
	"github.com/jmoiron/sqlx"
)

type T struct {
	Core *sqlx.Tx
	db   *Database
}

func (tx *T) getCore() (core StoragerCore) {
	return tx.Core
}
func (tx *T) getSQLChecker() (sqlChecker SQLChecker) {
	return tx.db.getSQLChecker()
}
func newTx(tx *sqlx.Tx, db *Database) *T {
	return &T{tx, db}
}

// TxResult
// tx.Commit() commit transaction
// tx.Rollback() rollback transaction , rollbackNoError = true
// tx.Error(err) rollback transaction , rollbackNoError = false, err = err
type TxResult struct {
	isCommit  bool
	withError error
}

func (T) Commit() TxResult {
	return TxResult{
		isCommit: true,
	}
}
func (T) Rollback() TxResult {
	return TxResult{
		isCommit: false,
	}
}
func (T) RollbackWithError(err error) TxResult {
	return TxResult{
		isCommit:  false,
		withError: xerr.WithStack(err),
	}
}

// Error same RollbackWithError
func (T) Error(err error) TxResult {
	return TxResult{
		isCommit:  false,
		withError: xerr.WithStack(err),
	}
}

// 给 TxResult 增加 Error 接口是为了避出现类似  tx.Rollback() 前面没有 return 的错误
func (result TxResult) Error() string {
	if result.withError != nil {
		return result.withError.Error()
	}
	if result.isCommit {
		return "goclub/sql: result commit"
	} else {
		return "goclub/sql: result rollback"
	}
}
func (db *Database) Begin(ctx context.Context, level sql.IsolationLevel, handle func(tx *T) TxResult) (rollbackNoError bool, err error) {
	return db.BeginOpt(ctx, sql.TxOptions{
		Isolation: level,
		ReadOnly:  false,
	}, handle)
}
func (db *Database) BeginOpt(ctx context.Context, opt sql.TxOptions, handle func(tx *T) TxResult) (rollbackNoError bool, err error) {
	coreTx, err := db.Core.BeginTxx(ctx, &opt)
	if err != nil {
		return
	}
	tx := newTx(coreTx, db)
	txResult := handle(tx)
	if txResult.isCommit {
		err = tx.Core.Commit()
		if err != nil {
			err = xerr.WithStack(err)
			return
		}
		return
	} else {
		err = tx.Core.Rollback()
		if err != nil {
			err = xerr.WithStack(err)
			return
		}
		if txResult.withError != nil {
			err = txResult.withError
			return
		}
		rollbackNoError = true
		return
	}
}

const (
	LevelDefault         sql.IsolationLevel = sql.LevelDefault
	LevelReadUncommitted sql.IsolationLevel = sql.LevelReadUncommitted
	LevelReadCommitted   sql.IsolationLevel = sql.LevelReadCommitted
	LevelWriteCommitted  sql.IsolationLevel = sql.LevelWriteCommitted
	LevelRepeatableRead  sql.IsolationLevel = sql.LevelRepeatableRead
	LevelSnapshot        sql.IsolationLevel = sql.LevelSnapshot
	LevelSerializable    sql.IsolationLevel = sql.LevelSerializable
	LevelLinearizable    sql.IsolationLevel = sql.LevelLinearizable
)

const (
	RC sql.IsolationLevel = sql.LevelReadCommitted
	RR sql.IsolationLevel = sql.LevelRepeatableRead
)
