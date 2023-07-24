package sq

import (
	"context"
	"database/sql"
	xerr "github.com/goclub/error"
	"github.com/jmoiron/sqlx"
)

type Transaction struct {
	Core *sqlx.Tx
	db   *Database
}

func (tx *Transaction) getCore() (core StoragerCore) {
	return tx.Core
}
func (tx *Transaction) getSQLChecker() (sqlChecker SQLChecker) {
	return tx.db.getSQLChecker()
}
func newTx(tx *sqlx.Tx, db *Database) *Transaction {
	return &Transaction{tx, db}
}

type TxResult struct {
	isCommit  bool
	withError error
}

func (Transaction) Commit() TxResult {
	return TxResult{
		isCommit: true,
	}
}
func (Transaction) Rollback() TxResult {
	return TxResult{
		isCommit: false,
	}
}
func (Transaction) RollbackWithError(err error) TxResult {
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

type BeginTransaction interface {
	BeginTransaction(ctx context.Context, level sql.IsolationLevel, handle func(tx *Transaction) TxResult) (err error)
}
type BeginTransactionOpt interface {
	BeginTransactionOpt(ctx context.Context, opt sql.TxOptions, handle func(tx *Transaction) TxResult) (err error)
}

func (db *Database) BeginTransaction(ctx context.Context, level sql.IsolationLevel, handle func(tx *Transaction) TxResult) (rollbackNoError bool, err error) {
	return db.BeginTransactionOpt(ctx, sql.TxOptions{
		Isolation: level,
		ReadOnly:  false,
	}, handle)
}
func (db *Database) BeginTransactionOpt(ctx context.Context, opt sql.TxOptions, handle func(tx *Transaction) TxResult) (rollbackNoError bool, err error) {
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
