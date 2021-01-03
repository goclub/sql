package sq

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"log"
	"reflect"
	"strings"
	"time"
)

type DB struct {
	Core *sqlx.DB
}
var createTimeField = []string{"CreatedAt","GMTCreate","CreatedTime",}
var updateTimeField = []string{"UpdatedAt", "GMTUpdate","UpdatedTime",}
var createAndUpdateTimeField = append(createTimeField, updateTimeField...)
func Open(driverName string, dataSourceName string) (db *DB, dbClose func() error, err error) {
	var coreDB *sqlx.DB
	coreDB ,err = sqlx.Open(driverName, dataSourceName)
	db = &DB{Core: coreDB,}
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
	ptr.BeforeCreate()
	qb := QB{
		Table: ptr,
	}
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		panic(errors.New("CreateModel(ctx, ptr) " + rType.String() + " must be ptr"))
	}
	elemValue := rValue.Elem()
	elemType := rType.Elem()
	for i:=0;i<elemType.NumField();i++ {
		itemType := elemType.Field(i)
		column, hasDBTag := itemType.Tag.Lookup("db")
		if !hasDBTag {continue}
		if column == "" {continue}
		sqTag := itemType.Tag.Get("sq")
		sqTags := strings.Split(sqTag, "|")
		shouldIgnoreInsert := false
		for _, tag := range sqTags {
			if tag == "ignore" { shouldIgnoreInsert = true ; break }
		}
		if shouldIgnoreInsert {continue}
		for _, field := range createAndUpdateTimeField {
			if itemType.Name == field {
				elemValue.Field(i).Set(reflect.ValueOf(time.Now()))
			}
		}
		qb.Insert = append(qb.Insert, Data{Column(column), elemValue.Field(i).Interface()})
	}
	query, values := qb.SQLInsert()
	result, err := db.Core.ExecContext(ctx, query, values...) ; if err != nil {
		return
	}
	err = ptr.AfterCreate(result) ; if err != nil {
		return
	}
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
	qb.Table = ptr
	qb.Limit = 1
	return db.QueryRowStructScan(ctx, ptr, qb)
}
func (db *DB) ModelList(ctx context.Context, modelSlicePtr interface{}, qb QB) error {
	return nil
}
func (db *DB) Update(ctx context.Context, qb QB) (err error) {
	query, values := qb.SQLUpdate()
	_, err = db.Core.ExecContext(ctx, query, values...)
	if err != nil {return err}
	return
}
func (db *DB) UpdateModel(ctx context.Context, ptr Model, updateData []Data, checkSQL ...string) (err error) {
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		panic(errors.New("UpdateModel(ctx, ptr) " + rType.String() + " must be ptr"))
	}
	elemValue := rValue.Elem()
	elemType := rType.Elem()
	idData := struct {
		HasID bool
		IDValue interface{}
	}{}
	for i:=0;i<elemType.NumField();i++ {
		itemType := elemType.Field(i)
		column, hasDBTag := itemType.Tag.Lookup("db")
		if !hasDBTag {continue}
		if column == "id" {
			idData.HasID = true
			idData.IDValue = elemValue.Field(i).Interface()
		}
		for _, field := range updateTimeField {
			if itemType.Name == field {
				updateTime := time.Now()
				elemValue.Field(i).Set(reflect.ValueOf(updateTime))
				sqTag := itemType.Tag.Get("sq")
				sqTags := strings.Split(sqTag, "|")
				var shouldIgnoreUpdate bool
				for _, tag := range sqTags {
					if tag == "ignore" { shouldIgnoreUpdate = true ; break }
				}
				if !shouldIgnoreUpdate {
					updateData = append(updateData, Data{Column(column), updateTime})
				}
			}
		}
		for _, data := range updateData {
			if column == data.Column.String() {
				elemValue.Field(i).Set(reflect.ValueOf(data.Value))
			}
		}
	}
	var where []Condition
	if idData.HasID {
		where = []Condition{{"id", Equal(idData.IDValue)}}
	} else {
		switch updateModeler := ptr.(type) {
		case UpdateModeler:
			where = updateModeler.UpdateModelWhere()
		default:
			return errors.New(elemType.Name() + " must has method UpdateModelWhere() sq.Condition or struct tag `db:\"id\"`")
		}
	}
	qb := QB{
		Table: ptr,
		Update: updateData,
		Where: where,
	}
	query, values := qb.SQLUpdate()
	_, err = db.Core.ExecContext(ctx, query, values...)
	if err != nil {return err}
	return
}
func (db *DB) SoftDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (err error) {
	return
}
func (db *DB) Relation(ctx context.Context, ptr Relation, qb QB, checkSQL ...string) (has bool, err error) {
	qb.Table = ptr
	qb.Limit = 1
	qb.Join = ptr.RelationJoin()
	return db.QueryRowStructScan(ctx, ptr, qb)
}
func (db *DB) RelationList(ctx context.Context, relationSlicePtr interface{}, qb QB, checkSQL ...string) (err error) {
	return
}
