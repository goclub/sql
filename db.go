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
	sqlChecker SQLChecker
}
func (db *Database) getCore() (core StoragerCore) {
	return db.Core
}
func (db *Database) getSQLChecker() (sqlChecker SQLChecker) {
	return db.sqlChecker
}
func (db *Database) SetSQLChecker(sqlChecker SQLChecker) {
	db.sqlChecker = sqlChecker
}
func Open(driverName string, dataSourceName string) (db *Database, dbClose func() error, err error) {
	var coreDatabase *sqlx.DB
	coreDatabase, err = sqlx.Open(driverName, dataSourceName)
	db = &Database{
		Core: coreDatabase,
		sqlChecker: &defaultSQLCheck{},
	}
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
var createTimeField = []string{"CreatedAt","GMTCreate","CreateTime",}
var updateTimeField = []string{"UpdatedAt", "GMTUpdate","UpdateTime",}
var createAndUpdateTimeField = append(createTimeField, updateTimeField...)
func (db *Database) Insert(ctx context.Context, qb QB) (result sql.Result, err error){
	return coreInsert(ctx, db, qb)
}
func (tx *Transaction) Insert(ctx context.Context, qb QB) (result sql.Result, err error){
	return coreInsert(ctx, tx, qb)
}
func coreInsert(ctx context.Context, storager Storager, qb QB) (result sql.Result, err error) {
	qb.SQLChecker = storager.getSQLChecker()
	return coreExec(ctx, storager, qb, Statement("").Enum().Insert)
}
// CreateModel
func (db *Database) CreateModel(ctx context.Context, ptr Model, checkSQL ...string) (err error) {
	return coreCreateModel(ctx, db, ptr)
}
func (tx *Transaction) CreateModel(ctx context.Context, ptr Model, checkSQL ...string) (err error) {
	return coreCreateModel(ctx, tx, ptr)
}

func coreCreateModel(ctx context.Context, storager Storager, ptr Model) (err error) {
	err = ptr.BeforeCreate() ; if err != nil {return}
	qb := QB{
		Table: ptr,
	}
	qb.SQLChecker = storager.getSQLChecker()
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		panic(errors.New("CreateModel(ctx, ptr) " + rType.String() + " must be ptr"))
	}
	elemValue := rValue.Elem()
	elemType := rType.Elem()
	eachField(elemValue, elemType, func(column string, fieldType reflect.StructField, fieldValue reflect.Value) {
		qb.Insert = append(qb.Insert, Insert{Column: Column(column), Value: fieldValue.Interface()})
	})
	raw := qb.SQLInsert()
	query, values := raw.Query, raw.Values
	result, err := storager.getCore().ExecContext(ctx, query, values...) ; if err != nil {
		return
	}
	err = ptr.AfterCreate(result) ; if err != nil {
		return
	}
	return
}
func eachField(elemValue reflect.Value, elemType reflect.Type, handle func(column string, fieldType reflect.StructField, fieldValue reflect.Value)) {
	for i:=0;i<elemType.NumField();i++ {
		fieldType := elemType.Field(i)
		fieldValue := elemValue.Field(i)
		// `db:"name"`
		column, hasDBTag := fieldType.Tag.Lookup("db")
		if fieldType.Anonymous == true {
			eachField(fieldValue, fieldValue.Type(), handle)
			continue
		}
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
		handle(column, fieldType, fieldValue)
	}
}
// QueryRowScan
func (db *Database) QueryRowScan(ctx context.Context, qb QB, desc ...interface{}) (has bool, err error) {
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryRowScan(ctx, db, qb, desc...)
}
func (tx *Transaction) QueryRowScan(ctx context.Context, qb QB, desc ...interface{}) (has bool, err error) {
	return coreQueryRowScan(ctx, tx, qb, desc...)
}
func coreQueryRowScan(ctx context.Context, storager Storager, qb QB, desc ...interface{}) (has bool, err error) {
	qb.SQLChecker = storager.getSQLChecker()
	qb.Limit = 1
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	row := storager.getCore().QueryRowxContext(ctx, query, values...)
	scanErr := row.Scan(desc...)
	has, err = CheckRowScanErr(scanErr) ; if err != nil {
		return
	}
	return
}
func (db *Database) SelectScan(ctx context.Context, qb QB, scan ScanFunc) (err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreSelectScan(ctx, db, qb, scan)
}
func (tx *Transaction) SelectScan(ctx context.Context, qb QB, scan ScanFunc) (error){
	return coreSelectScan(ctx, tx, qb, scan)
}
func coreSelectScan(ctx context.Context, storager Storager, qb QB, scan ScanFunc) (error) {
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	rows, err := storager.getCore().QueryxContext(ctx, query, values...) ; if err != nil {
		return  err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			return
		}
	}()
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
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryRowStructScan(ctx, db,ptr, qb)
}
func (tx *Transaction) QueryRowStructScan(ctx context.Context, ptr interface{}, qb QB)  (has bool, err error) {
	return coreQueryRowStructScan(ctx, tx, ptr, qb)
}
func coreQueryRowStructScan(ctx context.Context, storager Storager, ptr interface{}, qb QB)  (has bool, err error) {
	qb.SQLChecker = storager.getSQLChecker()
	qb.Limit = 1
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	row := storager.getCore().QueryRowxContext(ctx, query, values...)
	scanErr := row.StructScan(ptr)
	has, err = CheckRowScanErr(scanErr) ; if err != nil {
		return
	}
	return
}

func (db *Database) SelectSlice(ctx context.Context, slicePtr interface{}, qb QB) (err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreSelect(ctx, db, slicePtr, qb)
}
func (tx *Transaction) SelectSlice(ctx context.Context, slicePtr interface{}, qb QB) (err error){
	return coreSelect(ctx, tx, slicePtr, qb)
}
func coreSelect(ctx context.Context, storager Storager, slicePtr interface{}, qb QB) (err error) {
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	return storager.getCore().SelectContext(ctx, slicePtr, query, values...)
}
func (db *Database) Count(ctx context.Context, qb QB) (count int, err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreCount(ctx, db, qb)
}
func (tx *Transaction) Count(ctx context.Context, qb QB) (count int, err error){
	return coreCount(ctx, tx, qb)
}
func coreCount(ctx context.Context, storager Storager, qb QB) (count int, err error) {
	qb.SelectRaw = []Raw{{"COUNT(*)", nil}}
	qb.limitRaw = limitRaw{Valid: true, Limit: 0}
	var has bool
	has, err = coreQueryRowScan(ctx, storager, qb, &count);if err != nil {return }
	if has == false {
		raw := qb.SQLSelect()
		query := raw.Query
		panic(errors.New("goclub/sql: Count() " + query + "not found data"))
	}
	return
}
// if you need query data exited SELECT "has" FROM user WHERE id = ? better than SELECT count(*) FROM user where id = ?
func (db *Database) Has(ctx context.Context, qb QB) (has bool, err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreHas(ctx, db, qb)
}
func (tx *Transaction) Has(ctx context.Context, qb QB) (has bool, err error){
	return coreHas(ctx, tx, qb)
}
func coreHas(ctx context.Context, storager Storager, qb QB) (has bool, err error) {
	qb.SelectRaw = []Raw{{`1`, nil}}
	var i int
	return coreQueryRowScan(ctx, storager, qb, &i)
}
func (db *Database) Sum(ctx context.Context, column Column ,qb QB) (value sql.NullInt64, err error) {
	return coreSum(ctx, db, column, qb)
}
func (tx *Transaction) Sum(ctx context.Context, column Column ,qb QB) (value sql.NullInt64, err error) {
	return coreSum(ctx, tx, column, qb)
}
func coreSum(ctx context.Context, storager Storager,column Column ,qb QB) (value sql.NullInt64, err error) {
	qb.SelectRaw = []Raw{{"SUM(" + column.wrapField() + ")", nil}}
	_, err = coreQueryRowScan(ctx, storager, qb, &value) ; if err != nil {
		return
	}
	return
}
func (db *Database) QueryModel(ctx context.Context, ptr Model, qb QB) (has bool , err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryModel(ctx, db, ptr, qb)
}
func (tx *Transaction) QueryModel(ctx context.Context, ptr Model, qb QB) (has bool , err error){
	return coreQueryModel(ctx, tx, ptr, qb)
}
func coreQueryModel(ctx context.Context, storager Storager,ptr Model, qb QB) (has bool , err error) {
	qb.SQLChecker = storager.getSQLChecker()
	qb.Table = ptr
	qb.Limit = 1
	return coreQueryRowStructScan(ctx, storager, ptr, qb)
}
func (db *Database) QueryModelSlice(ctx context.Context, modelSlicePtr interface{}, qb QB) (err error) {
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryModelSlice(ctx, db, modelSlicePtr, qb)
}
func (tx *Transaction) QueryModelSlice(ctx context.Context, modelSlicePtr interface{}, qb QB) error {
	return coreQueryModelSlice(ctx, tx, modelSlicePtr, qb)
}
func coreQueryModelSlice(ctx context.Context, storager Storager, modelSlicePtr interface{}, qb QB) error {
	qb.SQLChecker = storager.getSQLChecker()
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
	err := storager.getCore().SelectContext(ctx, modelSlicePtr,query , values...) ; if err != nil {
		return err
	}
	return nil
}
func (db *Database) Update(ctx context.Context, qb QB) (result sql.Result, err error){
	return coreUpdate(ctx, db, qb)
}
func (tx *Transaction) Update(ctx context.Context, qb QB) (result sql.Result, err error){
	return coreUpdate(ctx, tx, qb)
}
func coreUpdate(ctx context.Context, storager Storager, qb QB) (result sql.Result, err error) {
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLUpdate()
	query, values := raw.Query, raw.Values
	result, err = storager.getCore().ExecContext(ctx, query, values...)
	if err != nil {return result, err}
	return
}

func (db *Database) UpdateModel(ctx context.Context, ptr Model, updateData []Update, where []Condition, checkSQL ...string) (result sql.Result, err error){
	return coreUpdateModel(ctx, db, ptr, updateData, where, checkSQL...)
}
func (tx *Transaction) UpdateModel(ctx context.Context, ptr Model, updateData []Update, where []Condition, checkSQL ...string) (result sql.Result, err error){
	return coreUpdateModel(ctx, tx, ptr, updateData, where, checkSQL...)
}
func coreUpdateModel(ctx context.Context, storager Storager, ptr Model, updateData []Update, where []Condition, checkSQL ...string) (result sql.Result, err error) {
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		panic(errors.New("UpdateModel(ctx, ptr) " + rType.String() + " must be ptr"))
	}
	elemValue := rValue.Elem()
	elemType := rType.Elem()
	primaryIDInfo := struct {
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
			primaryIDInfo.HasID = true
			primaryIDInfo.IDValue = fieldValue.Interface()
		}

		//  updated time.Time
		for _, timeField := range updateTimeField {
			if fieldType.Name == timeField {
				setTimeNow(fieldValue, fieldType)
				// UpdatedAt time.Time `sq:"ignore"`
				shouldIgnore := Tag{fieldType.Tag.Get("sq")}.IsIgnore()
				if !shouldIgnore {
					updateData = append(updateData, Update{
						Column: Column(column),
						Value: fieldValue.Interface(),
					})
				}
			}
		}
		for dataIndex, data := range updateData {
			if len(data.Column) == 0  && column == data.Column.String() {
					if data.OnUpdated == nil {
						updateData[dataIndex].OnUpdated = func() error {
							fieldValue.Set(reflect.ValueOf(data.Value))
							return nil
						}
					}
			}
		}
	}
	primaryKeyWhere, err := primaryKeyWhere(ptr, primaryIDInfo, elemType.Name()) ; if err != nil {
		return
	}
	wheres := append(primaryKeyWhere, where...)
	qb := QB{
		Table: ptr,
		Update: updateData,
		Where: wheres,
	}
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLUpdate()
	query, values := raw.Query, raw.Values
	result, err = storager.getCore().ExecContext(ctx, query, values...)
	if err != nil {return result, err}
	for _, data := range updateData {
		if data.OnUpdated != nil {
			updatedErr := data.OnUpdated() ; if updatedErr != nil {
				return result, updatedErr
			}
		}
	}
	return
}
func (db *Database) HardDelete(ctx context.Context, qb QB, checkSQL ...string) (result sql.Result, err error) {
	return coreHardDelete(ctx, db, qb, checkSQL...)
}
func (tx *Transaction) HardDelete(ctx context.Context, qb QB, checkSQL ...string) (result sql.Result, err error) {
	return coreHardDelete(ctx, tx, qb, checkSQL...)
}
func coreHardDelete(ctx context.Context, storager Storager, qb QB, checkSQL ...string) (result sql.Result, err error) {
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLDelete()
	return storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
}
func (db *Database) HardDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (result sql.Result, err error){
	return coreHardDeleteModel(ctx,db, ptr, checkSQL...)
}
func (tx *Transaction) HardDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (result sql.Result, err error){
	return coreHardDeleteModel(ctx, tx, ptr, checkSQL...)
}
func coreHardDeleteModel(ctx context.Context, storager Storager, ptr Model, checkSQL ...string) (result sql.Result, err error) {
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		panic(errors.New("UpdateModel(ctx, ptr) " + rType.String() + " must be ptr"))
	}
	elemValue := rValue.Elem()
	elemType := rType.Elem()
	primaryIDInfo := struct {
		HasID bool
		IDValue interface{}
	}{}
	for i:=0;i<elemType.NumField();i++ {
		fieldType := elemType.Field(i)
		fieldValue := elemValue.Field(i)
		column, hasDBTag := fieldType.Tag.Lookup("db")
		if !hasDBTag {
			continue
		}
		// find primary id
		if column == "id" {
			primaryIDInfo.HasID = true
			primaryIDInfo.IDValue = fieldValue.Interface()
		}
	}
	primaryKeyWhere, err := primaryKeyWhere(ptr, primaryIDInfo, elemType.Name()) ; if err != nil {
		return
	}
	qb := QB{
		Table: ptr,
		Where: primaryKeyWhere,
		Limit: 1,
	}
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLDelete()
	return storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
}
func (db *Database) SoftDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (result sql.Result, err error){
	return coreSoftDeleteModel(ctx, db, ptr, checkSQL...)
}
func (tx *Transaction) SoftDeleteModel(ctx context.Context, ptr Model, checkSQL ...string) (result sql.Result, err error){
	return coreSoftDeleteModel(ctx, tx, ptr, checkSQL...)
}
func coreSoftDeleteModel(ctx context.Context, storager Storager, ptr Model, checkSQL ...string) (result sql.Result, err error) {
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		panic(errors.New("UpdateModel(ctx, ptr) " + rType.String() + " must be ptr"))
	}
	elemValue := rValue.Elem()
	elemType := rType.Elem()
	primaryIDInfo := struct {
		HasID bool
		IDValue interface{}
	}{}
	for i:=0;i<elemType.NumField();i++ {
		fieldType := elemType.Field(i)
		fieldValue := elemValue.Field(i)
		column, hasDBTag := fieldType.Tag.Lookup("db")
		if !hasDBTag {
			continue
		}
		// find primary id
		if column == "id" {
			primaryIDInfo.HasID = true
			primaryIDInfo.IDValue = fieldValue.Interface()
		}
	}
	primaryKeyWhere, err := primaryKeyWhere(ptr, primaryIDInfo, elemType.Name()) ; if err != nil {
		return
	}
	qb := QB{
		Table: ptr,
		Where: primaryKeyWhere,
		Update: []Update{{Raw:ptr.SoftDeleteWhere(),}},
		Limit: 1,
	}
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLUpdate()
	return storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)

}
func (db *Database) QueryRelation(ctx context.Context, ptr Relation, qb QB, checkSQL ...string) (has bool, err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryRelation(ctx, db, ptr, qb, checkSQL...)
}
func (tx *Transaction) QueryRelation(ctx context.Context, ptr Relation, qb QB, checkSQL ...string) (has bool, err error){
	return coreQueryRelation(ctx, tx, ptr, qb, checkSQL...)
}
func coreQueryRelation(ctx context.Context, storager Storager, ptr Relation, qb QB, checkSQL ...string) (has bool, err error) {
	qb.SQLChecker = storager.getSQLChecker()
	qb.Table = ptr
	qb.Limit = 1
	qb.Join = ptr.RelationJoin()
	return coreQueryRowStructScan(ctx, storager, ptr, qb)
}
func (db *Database) QueryRelationSlice(ctx context.Context, relationSlicePtr interface{}, qb QB, checkSQL ...string) (err error) {
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryRelationSlice(ctx, db, relationSlicePtr, qb, checkSQL...)
}
func (tx *Transaction) QueryRelationSlice(ctx context.Context, relationSlicePtr interface{}, qb QB, checkSQL ...string) (err error) {
	return coreQueryRelationSlice(ctx, tx, relationSlicePtr, qb, checkSQL...)
}
func coreQueryRelationSlice(ctx context.Context, storager Storager, relationSlicePtr interface{}, qb QB, checkSQL ...string) (err error) {
	qb.SQLChecker = storager.getSQLChecker()
	ptrType := reflect.TypeOf(relationSlicePtr)
	if ptrType.Kind() != reflect.Ptr {
		panic(errors.New("goclub/sql: " + ptrType.String() + "not pointer"))
	}
	elemType := ptrType.Elem()
	reflectItemValue := reflect.MakeSlice(elemType, 1,1).Index(0)
	tablerInterface := reflectItemValue.Interface().(Relation)
	qb.Table = tablerInterface
	qb.Join = tablerInterface.RelationJoin()
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	err = storager.getCore().SelectContext(ctx, relationSlicePtr,query , values...) ; if err != nil {
		return err
	}
	return
}
func (db *Database) Exec(ctx context.Context, qb QB, statement Statement) (result sql.Result, err error){
	return coreExec(ctx, db, qb, statement)
}
func (tx *Transaction) Exec(ctx context.Context, qb QB, statement Statement) (result sql.Result, err error){
	return coreExec(ctx, tx, qb, statement)
}
func coreExec(ctx context.Context, storager Storager, qb QB, statement Statement) (result sql.Result, err error) {
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQL(statement)
	result, err = storager.getCore().ExecContext(ctx, raw.Query, raw.Values...) ; if err != nil {
		return
	}
	return
}