package sq

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Transaction struct {
	Core *sqlx.Tx
	sqlChecker SQLChecker
}
func (tx *Transaction) getCore() (core StoragerCore) {
	return tx.Core
}
func (tx *Transaction) getSQLChecker() (sqlChecker SQLChecker) {
	return tx.sqlChecker
}
func newTx(tx *sqlx.Tx, sqlChecker SQLChecker) *Transaction {
	return &Transaction{tx, sqlChecker}
}

type TxResult struct {
	isCommit bool
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
		isCommit: false,
		withError: err,
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

var ErrTransactionIsRollback = errors.New("goclub/sql: transaction rollback")
func (db *Database) BeginTransaction(ctx context.Context, level sql.IsolationLevel, handle func (tx *Transaction) TxResult) (err error) {
	return db.BeginTransactionOpt(ctx, sql.TxOptions{
		Isolation: level,
		ReadOnly: false,
	}, handle)
}
func (db *Database) BeginTransactionOpt(ctx context.Context, opt sql.TxOptions, handle func (tx *Transaction) TxResult) ( err error) {
	coreTx, err := db.Core.BeginTxx(ctx, &opt) ; if err != nil {
		return
	}
	tx := newTx(coreTx, db.sqlChecker)
	txResult := handle(tx)
	if txResult.isCommit {
		err = tx.Core.Commit() ; if err != nil {
			return
		}
		return
	} else {
		err = tx.Core.Rollback() ; if err != nil {
			return err
		}
		if txResult.withError != nil {
			return txResult.withError
		}
		return ErrTransactionIsRollback
	}
}
