package sq

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"log"
	"reflect"
)

type DB struct {
	Core *sqlx.DB
}
var createTimeField = []string{"CreatedAt","GMTCreate","CreateTime",}
var updateTimeField = []string{"UpdatedAt", "GMTUpdate","UpdateTime",}
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
	err = ptr.BeforeCreate() ; if err != nil {return}
	qb := QB{
		Table: ptr,
		Debug:true,
	}
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		panic(errors.New("CreateModel(ctx, ptr) " + rType.String() + " must be ptr"))
	}
	elemValue := rValue.Elem()
	elemType := rType.Elem()
	for i:=0;i<elemType.NumField();i++ {
		fieldType := elemType.Field(i)
		fieldValue := elemValue.Field(i)
		// `db:"name"`
		column, hasDBTag := fieldType.Tag.Lookup("db")
		if !hasDBTag {continue}
		if column == "" {continue}

		// `sq:"ignore"`
		shouldIgnoreInsert := Tag{fieldType.Tag.Get("sq")}.IsIgnore()
		if shouldIgnoreInsert {continue}
		// created updated time.Time
		for _, timeField := range createAndUpdateTimeField {
			if fieldType.Name == timeField {
				setTimeNow(fieldValue, fieldType)
			}
		}
		qb.Insert = append(qb.Insert, Data{Column(column), fieldValue.Interface()})
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
	scanErr := row.Scan(desc...) ; if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return false, nil
		} else {
			return false, scanErr
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
	scanErr := row.StructScan(ptr) ; if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return false, nil
		} else {
			return false, scanErr
		}
	} else {
		has = true
	}
	return
}
func (db *DB) Count(ctx context.Context, qb QB) (count int, err error) {
	qb.SelectRaw = []QueryValues{{"COUNT(*)", nil}}
	qb.limitRaw = limitRaw{Valid: true, Limit: 0}
	var has bool
	has, err = db.QueryRowScan(ctx, qb, &count);if err != nil {return }
	if has == false {
		query, _ := qb.SQLSelect()
		panic(errors.New("goclub/sql: Count() " + query + "not found data"))
	}
	return
}
func (db *DB) QueryModel(ctx context.Context, ptr Model, qb QB) (has bool , err error) {
	qb.Table = ptr
	qb.Limit = 1
	return db.QueryRowStructScan(ctx, ptr, qb)
}
func (db *DB) QueryModelList(ctx context.Context, modelSlicePtr interface{}, qb QB) error {
	elemType := reflect.TypeOf(modelSlicePtr).Elem()
	reflectItemValue := reflect.MakeSlice(elemType, 1,1).Index(0)
	tablerInterface := reflectItemValue.Interface().(Tabler)
	qb.Table = tablerInterface
	query, values := qb.SQLSelect()
	err := db.Core.SelectContext(ctx, modelSlicePtr,query , values...) ; if err != nil {
		return err
	}
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
		fieldType := elemType.Field(i)
		fieldValue := elemValue.Field(i)
		column, hasDBTag := fieldType.Tag.Lookup("db")
		if !hasDBTag {continue}

		// find primary id
		if column == "id" {
			idData.HasID = true
			idData.IDValue = fieldValue.Interface()
		}

		//  updated time.Time
		for _, timeField := range updateTimeField {
			if fieldType.Name == timeField {
				setTimeNow(fieldValue, fieldType)
				// UpdatedAt time.Time `sq:"ignore"`
				shouldIgnore := Tag{fieldType.Tag.Get("sq")}.IsIgnore()
				if !shouldIgnore {
					updateData = append(updateData, Data{Column(column), fieldValue.Interface()})
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
func (db *DB) QueryRelation(ctx context.Context, ptr Relation, qb QB, checkSQL ...string) (has bool, err error) {
	qb.Table = ptr
	qb.Limit = 1
	qb.Join = ptr.RelationJoin()
	return db.QueryRowStructScan(ctx, ptr, qb)
}
func (db *DB) QueryRelationList(ctx context.Context, relationSlicePtr interface{}, qb QB, checkSQL ...string) (err error) {
	return
}
