package sq

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"log"
	"reflect"
)

type Database struct {
	Core *sqlx.DB
}
var createTimeField = []string{"CreatedAt","GMTCreate","CreateTime",}
var updateTimeField = []string{"UpdatedAt", "GMTUpdate","UpdateTime",}
var createAndUpdateTimeField = append(createTimeField, updateTimeField...)
func Open(driverName string, dataSourceName string) (db *Database, dbClose func() error, err error) {
	var coreDatabase *sqlx.DB
	coreDatabase, err = sqlx.Open(driverName, dataSourceName)
	db = &Database{Core: coreDatabase,}
	if err != nil && coreDatabase != nil {
		dbClose = coreDatabase.Close
	} else {
		dbClose = func() error { return nil}
	}
	if err != nil {
		return
	}
	return
}
func (db *Database) Close() error {
	if db.Core != nil {
		return db.Core.Close()
	}
	log.Print("Database is nil,maybe you forget sq.Open()")
	return nil
}
func (db *Database) CreateModel(ctx context.Context, ptr Model) (err error) {
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
		// `Database:"name"`
		column, hasDatabaseTag := fieldType.Tag.Lookup("Database")
		if !hasDatabaseTag {continue}
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
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	result, err := db.Core.ExecContext(ctx, query, values...) ; if err != nil {
		return
	}
	err = ptr.AfterCreate(result) ; if err != nil {
		return
	}
	return
}
func CheckRowScanErr(scanErr error) (has bool, err error) {
	if scanErr != nil {
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
func (db *Database) QueryRowScan(ctx context.Context, qb QB, desc ...interface{}) (has bool, err error) {
	return coreQueryRowScan(db.Core, ctx, qb, desc...)
}
func (tx *Tx) QueryRowScan(ctx context.Context, qb QB, desc ...interface{}) (has bool, err error) {
	return coreQueryRowScan(tx.Core, ctx, qb, desc...)
}
func coreQueryRowScan(storager Storager, ctx context.Context, qb QB, desc ...interface{}) (has bool, err error) {
	qb.Limit = 1
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	row := storager.QueryRowx(query, values...)
	scanErr := row.Scan(desc...)
	has, err = CheckRowScanErr(scanErr) ; if err != nil {
		return
	}
	return
}
func (db *Database) SelectScan(ctx context.Context,qb QB, scan ScanFunc) (error) {
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	rows, err := db.Core.Queryx(query, values...) ; if err != nil {
		return  err
	}
	defer rows.Close()
	for rows.Next() {
		err := scan(rows) ; if err != nil {
			return err
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return rowsErr
	}
	return nil
}
func (db *Database) QueryRowStructScan(ctx context.Context, ptr interface{}, qb QB)  (has bool, err error) {
	return coreQueryRowStructScan(db.Core, ctx, ptr, qb)
}
func (tx *Tx) QueryRowStructScan(ctx context.Context, ptr interface{}, qb QB)  (has bool, err error) {
	return coreQueryRowStructScan(tx.Core, ctx, ptr, qb)
}
func coreQueryRowStructScan(core Storager, ctx context.Context, ptr interface{}, qb QB)  (has bool, err error) {
	qb.Limit = 1
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	row := core.QueryRowx(query, values...)
	scanErr := row.StructScan(ptr)
	has, err = CheckRowScanErr(scanErr) ; if err != nil {
		return
	}
	return
}

func (db *Database) Select(ctx context.Context, slicePtr interface{}, qb QB) (err error) {
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	return db.Core.SelectContext(ctx, slicePtr, query, values...)
}
func (db *Database) Count(ctx context.Context, qb QB) (count int, err error) {
	qb.SelectRaw = []Raw{{"COUNT(*)", nil}}
	qb.limitRaw = limitRaw{Valid: true, Limit: 0}
	var has bool
	has, err = db.QueryRowScan(ctx, qb, &count);if err != nil {return }
	if has == false {
		raw := qb.SQLSelect()
		query := raw.Query
		panic(errors.New("goclub/sql: Count() " + query + "not found data"))
	}
	return
}
func (db *Database) QueryModel(ctx context.Context, ptr Model, qb QB) (has bool , err error) {
	qb.Table = ptr
	qb.Limit = 1
	return db.QueryRowStructScan(ctx, ptr, qb)
}
func (db *Database) QueryModelList(ctx context.Context, modelSlicePtr interface{}, qb QB) error {
	ptrType := reflect.TypeOf(modelSlicePtr)
	if ptrType.Kind() != reflect.Ptr {
		panic(errors.New("goclub/sql: " + ptrType.String() + "not pointer"))
	}
	elemType := ptrType.Elem()
	reflectItemValue := reflect.MakeSlice(elemType, 1,1).Index(0)
	tablerInterface := reflectItemValue.Interface().(Tabler)
	qb.Table = tablerInterface
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	err := db.Core.SelectContext(ctx, modelSlicePtr,query , values...) ; if err != nil {
		return err
	}
	return nil
}
func (db *Database) Update(ctx context.Context, qb QB) (err error) {
	raw := qb.SQLUpdate()
	query, values := raw.Query, raw.Values
	_, err = db.Core.ExecContext(ctx, query, values...)
	if err != nil {return err}
	return
}
func (db *Database) UpdateModel(ctx context.Context, ptr Model, updateData []Data, checkSQL ...string) (err error) {
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
		column, hasDatabaseTag := fieldType.Tag.Lookup("Database")
		if !hasDatabaseTag {continue}

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
			return errors.New(elemType.Name() + " must has method UpdateModelWhere() sq.Condition or struct tag `Database:\"id\"`")
		}
	}
	qb := QB{
		Table: ptr,
		Update: updateData,
		Where: where,
	}
	raw := qb.SQLUpdate()
	query, values := raw.Query, raw.Values
	_, err = db.Core.ExecContext(ctx, query, values...)
	if err != nil {return err}
	return
}
func (db *Database) SoftDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (err error) {
	return
}
func (db *Database) QueryRelation(ctx context.Context, ptr Relation, qb QB, checkSQL ...string) (has bool, err error) {
	qb.Table = ptr
	qb.Limit = 1
	qb.Join = ptr.RelationJoin()
	return db.QueryRowStructScan(ctx, ptr, qb)
}
func (db *Database) QueryRelationList(ctx context.Context, relationSlicePtr interface{}, qb QB, checkSQL ...string) (err error) {
	return
}
func (db *Database) Exec(ctx context.Context, qb QB) (result sql.Result, err error) {

	return
}