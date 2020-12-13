package sq

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"log"
)

type DB struct {
	Core *sqlx.DB
}
func Open(driverName string, dataSourceName string) (db *DB, dbClose func() error, err error) {
	var coreDB *sqlx.DB
	coreDB ,err = sqlx.Open(driverName, dataSourceName)
	db = &DB{
		Core: coreDB,
	}
	if err != nil && coreDB != nil {
		dbClose = coreDB.Close
	} else {
		dbClose = func() error { return nil}
	}
	return
}
func (db *DB) Close() error {
	if db.Core != nil {
		return db.Core.Close()
	}
	log.Print("db is nil,maybe you forget sq.Open()")
	return nil
}
func (db *DB) One(ctx context.Context, ptr Model, qb QB) (has bool , error error) {
	qb = qb.BindModel(ptr)
	return db.QueryRowStructScan(ctx, ptr, qb)
}
func (db *DB) QueryRowScan(ctx context.Context, qb QB, desc ...interface{}) (has bool, err error) {
	qb.Limit = 1
	query, values := qb.ToSelect()
	row := db.Core.QueryRowx(query, values...)
	err = row.Scan(desc...) ; if err != nil {
		if err == sql.ErrNoRows {
			has = false
		} else {
			return
		}
	} else {
		has = true
	}
	return
}
func (db *DB) QueryRowStructScan(ctx context.Context, ptr interface{}, qb QB)  (has bool, err error) {
	qb.Limit = 1
	query, values := qb.ToSelect()
	row := db.Core.QueryRowx(query, values...)
	err = row.StructScan(ptr) ; if err != nil {
		if err == sql.ErrNoRows {
			has = false
		} else {
			return
		}
	} else {
		has = true
	}
	return
}
func (db *DB) Count(ctx context.Context, ptr Model, qb QB) (count int, err error) {
	qb.Select = []Column{"COUNT(*)"}
	qb = qb.BindModel(ptr)
	var has bool
	has, err = db.QueryRowScan(ctx, qb, &count);if err != nil {return }
	if has == false {
		query, _ := qb.ToSelect()
		panic(errors.New("goclub/sql: Count() " + query + "not found data"))
	}
	return
}
func (db *DB) List(ctx context.Context, slicePtr interface{}, qb QB) error {
	return nil
}
func (db *DB) Select(ctx context.Context, slicePtr interface{}, qb QB) error {
	return nil
}
func (db *DB) Exec(ctx context.Context, query string, values []interface{}) (result sql.Result, err error) {
	return
}