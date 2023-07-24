package sq

import (
	"database/sql"
	xerr "github.com/goclub/error"
)

type Result struct {
	core sql.Result
}

func (r Result) LastInsertId() (id int64, err error) {
	id, err = r.core.LastInsertId()
	if err != nil {
		err = xerr.WithStack(err)
		return
	}
	return
}
func (r Result) LastInsertUint64Id() (id uint64, err error) {
	var int64id int64
	int64id, err = r.LastInsertId()
	if err != nil {
		return
	}
	if int64id < 0 {
		err = xerr.New("goclub/sql: sq.Result{}.LastInsertUint64Id() (id, err) id less than 0")
		return
	}
	id = uint64(int64id)
	return
}
func (r Result) RowsAffected() (rowsAffected int64, err error) {
	rowsAffected, err = r.core.RowsAffected()
	if err != nil {
		err = xerr.WithStack(err)
		return
	}
	return
}

func RowsAffected(result Result, execErr error) (affected int64, err error) {
	if execErr != nil {
		err = execErr
		return
	}
	return result.RowsAffected()
}
