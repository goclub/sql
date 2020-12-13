package sq

import (
	"context"
	"database/sql"
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
	return
}
func (db *DB) Scan(ctx context.Context, qb QB, desc ...interface{})  error {
	return nil
}
func (db *DB) ScanStruct(ctx context.Context, ptr interface{}, qb QB)  (has bool, err error) {
	return
}
func (db *DB) List(ctx context.Context, slicePtr interface{}, qb QB) error {
	return nil
}
func (db *DB) Select(ctx context.Context, slicePtr interface{}, qb QB) error {
	return nil
}
func (db *DB) Count(ctx context.Context, ptr Model, qb QB) (count int, err error) {
	return
}
func (db *DB) Exec(ctx context.Context, query string, values []interface{}) (result sql.Result, err error) {
	return
}