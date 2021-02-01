package sq

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type Transaction struct {
	Core *sqlx.Tx
}
func newTx(tx *sqlx.Tx) *Transaction {
	return &Transaction{tx}
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

func (db *Database) Transaction(ctx context.Context, handle func (tx *Transaction) TxResult) (err error) {
	return db.TransactionOpts(ctx, handle, nil)
}
func (db *Database) TransactionOpts(ctx context.Context, handle func (tx *Transaction) TxResult, opts *sql.TxOptions) (err error) {
	coreTx, err := db.Core.BeginTxx(ctx, opts) ; if err != nil {
		return
	}
	tx := newTx(coreTx)
	txResult := handle(tx)
	if txResult.isCommit {
		err = tx.Core.Commit() ; if err != nil {
			return
		}
	} else {
		err = tx.Core.Rollback() ; if err != nil {
			return
		}
		if txResult.withError != nil {
			return txResult.withError
		}
	}
	return
}