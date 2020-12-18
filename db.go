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
func (db *DB) Exec(ctx context.Context, qb QB) (result sql.Result, err error) {
	return
}
func (db *DB) CreateModel(ctx context.Context, ptr Model, checkSQL ...string) (err error) {
	return
}
func (db *DB) MultiCreateModel(ctx context.Context, modelSlicePtr interface{}, checkSQL ...string) (err error) {
	return
}
func (db *DB) QueryRowScan(ctx context.Context, qb QB, desc ...interface{}) (has bool, err error) {
	qb.Limit = 1
	query, values := qb.SQLSelect()
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
	query, values := qb.SQLSelect()
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
func (db *DB) Count(ctx context.Context, qb QB) (count int, err error) {
	qb.Select = []Column{"COUNT(*)"}
	var has bool
	has, err = db.QueryRowScan(ctx, qb, &count);if err != nil {return }
	if has == false {
		query, _ := qb.SQLSelect()
		panic(errors.New("goclub/sql: Count() " + query + "not found data"))
	}
	return
}
func (db *DB) Model(ctx context.Context, ptr Model, qb QB) (has bool , err error) {
	qb = qb.BindModel(ptr)
	return db.QueryRowStructScan(ctx, ptr, qb)
}
func (db *DB) ModelList(ctx context.Context, modelSlicePtr interface{}, qb QB) error {
	return nil
}
func (db *DB) Update(ctx context.Context, qb QB) (err error) {
	return
}
func (db *DB) UpdateModel(ctx context.Context, model Model, data map[Column]interface{}, checkSQL ...string) (err error) {
	return
}
func (db *DB) SoftDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (err error) {
	return
}
func (db *DB) Relation(ctx context.Context, ptr Relation, qb QB, checkSQL ...string) (err error) {
	return
}
func (db *DB) RelationList(ctx context.Context, relationSlicePtr interface{}, qb QB, checkSQL ...string) (err error) {
	return
}
