package sq

import (
	"context"
	"database/sql"
	xerr "github.com/goclub/error"
	"github.com/jmoiron/sqlx"
	"log"
	"reflect"
	"strings"
)

type Database struct {
	Core *sqlx.DB
	sqlChecker SQLChecker
}
func (db *Database) Ping(ctx context.Context) error {
	return db.Core.PingContext(ctx)
}
func (db *Database) getCore() (core StoragerCore) {
	return db.Core
}
func (db *Database) getSQLChecker() (sqlChecker SQLChecker) {
	return db.sqlChecker
}
func (db *Database) SetSQLChecker(sqlChecker SQLChecker)  {
	db.sqlChecker = sqlChecker
}


func Open(driverName string, dataSourceName string) (db *Database, dbClose func() error, err error) {
	var coreDatabase *sqlx.DB
	coreDatabase, err = sqlx.Open(driverName, dataSourceName)
	db = &Database{
		Core: coreDatabase,
		sqlChecker: &DefaultSQLChecker{},
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
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SQLChecker = storager.getSQLChecker()
	return coreExecQB(ctx, storager, qb, StatementInsert)
}

func (db *Database) InsertModel(ctx context.Context, ptr Model, qb QB) (result sql.Result, err error) {
	return coreInsertModel(ctx, db, ptr,  qb)
}
func (tx *Transaction) InsertModel(ctx context.Context, ptr Model, qb QB) (result sql.Result, err error) {
	return coreInsertModel(ctx, tx, ptr, qb)
}

func coreInsertModel(ctx context.Context, storager Storager, ptr Model, qb QB) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	err = ptr.BeforeCreate() ; if err != nil {return}
	if qb.From != nil {
		log.Print("InsertModelBaseOnQB(ctx, qb, model) qb.From need be nil")
	}
	qb.From = ptr
	qb.SQLChecker = storager.getSQLChecker()
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		return result, xerr.New("InsertModel(ctx, ptr) " + rType.String() + " must be ptr")
	}
	elemValue := rValue.Elem()
	elemType := rType.Elem()
	if len(qb.Insert) == 0 && len(qb.InsertMultiple.Column) == 0 {
		insertEachField(elemValue, elemType, func(column string, fieldType reflect.StructField, fieldValue reflect.Value) {
			qb.Insert = append(qb.Insert, Insert{Column: Column(column), Value: fieldValue.Interface()})
		})
	}
	raw := qb.SQLInsert()
	query, values := raw.Query, raw.Values
	result, err = storager.getCore().ExecContext(ctx, query, values...) ; if err != nil {
		return
	}
	err = ptr.AfterCreate(result) ; if err != nil {
		return
	}
	return
}
func insertEachField(elemValue reflect.Value, elemType reflect.Type, handle func(column string, fieldType reflect.StructField, fieldValue reflect.Value)) {
	for i:=0;i<elemType.NumField();i++ {
		fieldType := elemType.Field(i)
		fieldValue := elemValue.Field(i)
		// `db:"name"`
		column, hasDBTag := fieldType.Tag.Lookup("db")
		if fieldType.Anonymous == true {
			insertEachField(fieldValue, fieldValue.Type(), handle)
			continue
		}
		if !hasDBTag {continue}
		if column == "" {continue}
		// `sq:"ignoreInsert"`
		shouldIgnoreInsert := Tag{fieldType.Tag.Get("sq")}.IsIgnoreInsert()
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
func (db *Database) QueryRowScan(ctx context.Context, qb QB, desc []interface{}) (has bool, err error) {
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryRowScan(ctx, db, qb, desc)
}
func (tx *Transaction) QueryRowScan(ctx context.Context, qb QB, desc []interface{}) (has bool, err error) {
	return coreQueryRowScan(ctx, tx, qb, desc)
}
func coreQueryRowScan(ctx context.Context, storager Storager, qb QB, desc []interface{}) (has bool, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
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
func (db *Database) QuerySliceScaner(ctx context.Context, qb QB, scan Scaner) (err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQuerySliceScaner(ctx, db, qb, scan)
}
func (tx *Transaction) QuerySliceScaner(ctx context.Context, qb QB, scan Scaner)  (error){
	return coreQuerySliceScaner(ctx, tx, qb, scan)
}
func coreQuerySliceScaner(ctx context.Context, storager Storager, qb QB, scan Scaner) (err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
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
func (db *Database) Query(ctx context.Context, ptr Tabler, qb QB)  (has bool, err error) {
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQuery(ctx, db,ptr, qb)
}
func (tx *Transaction) Query(ctx context.Context, ptr Tabler, qb QB)  (has bool, err error) {
	return coreQuery(ctx, tx, ptr, qb)
}

// func (db *Database) QueryModel(ctx context.Context, ptr Model, qb QB)  (has bool, err error) {
// 	qb.Where = ptr.PrimaryKey()
// 	err = qb.mustInTransaction() ; if err != nil {return}
// 	return coreQuery(ctx, db, ptr, qb)
// }
// func (tx *Transaction) QueryModel(ctx context.Context, ptr Model, qb QB)  (has bool, err error) {
// 	qb.Where = ptr.PrimaryKey()
// 	return coreQuery(ctx, tx, ptr, qb)
// }

func coreQuery(ctx context.Context, storager Storager, ptr Tabler, qb QB)  (has bool, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SQLChecker = storager.getSQLChecker()
	qb.Limit = 1
	qb.From = ptr
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	row := storager.getCore().QueryRowxContext(ctx, query, values...)
	scanErr := row.StructScan(ptr)
	has, err = CheckRowScanErr(scanErr) ; if err != nil {
		return
	}
	return
}

func (db *Database) QuerySlice(ctx context.Context, slicePtr interface{}, qb QB) (err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQuerySlice(ctx, db, slicePtr, qb)
}
func (tx *Transaction) QuerySlice(ctx context.Context, slicePtr interface{}, qb QB) (err error){
	return coreQuerySlice(ctx, tx, slicePtr, qb)
}
func coreQuerySlice(ctx context.Context, storager Storager, slicePtr interface{}, qb QB) (err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SQLChecker = storager.getSQLChecker()
	ptrType := reflect.TypeOf(slicePtr)
	if ptrType.Kind() != reflect.Ptr {
		return xerr.New("goclub/sql: " + ptrType.String() + "not pointer")
	}
	if qb.From == nil && qb.FromRaw.TableName.Query  == "" && qb.Raw.Query == "" {
		elemType := ptrType.Elem()
		reflectItemValue := reflect.MakeSlice(elemType, 1,1).Index(0)
		if reflectItemValue.CanAddr() {
			reflectItemValue = reflectItemValue.Addr()
		}
		tablerInterface := reflectItemValue.Interface().(Tabler)
		qb.From = tablerInterface
	}
	if qb.From == nil && qb.FromRaw.TableName.Query  == "" && qb.Raw.Query == "" {
		// 如果设置了 qb.Form 但没有设置 qb.Select 可能会导致 select * ,这种情况在代码已经在线上运行时但是表变动了时会很危险
		if len(qb.Select) == 0 && len(qb.SelectRaw) == 0 {
			err = xerr.New("goclub/sql: QuerySlice(ctx, slice, qb) if qb.Form/qb.FromRaw/qb.Raw not zero value, then qb.Select or qb.SelectRaw can not be nil, or you can set qb.Form/qb.FromRaw/ be nil")
			return
		}
	}
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	return storager.getCore().SelectContext(ctx, slicePtr, query, values...)
}
func (db *Database) Count(ctx context.Context, qb QB) (count uint64, err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreCount(ctx, db, qb)
}
func (tx *Transaction) Count(ctx context.Context, qb QB) (count uint64, err error){
	return coreCount(ctx, tx, qb)
}
func coreCount(ctx context.Context, storager Storager, qb QB) (count uint64, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SelectRaw = []Raw{{"COUNT(*)", nil}}
	qb.limitRaw = limitRaw{Valid: true, Limit: 0}
	var has bool
	has, err = coreQueryRowScan(ctx, storager, qb, []interface{}{&count});if err != nil {return }
	if has == false {
		raw := qb.SQLSelect()
		query := raw.Query
		return 0, xerr.New("goclub/sql: Count() " + query + "not found data")
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
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SelectRaw = []Raw{{`1`, nil}}
	var i int
	return coreQueryRowScan(ctx, storager, qb, []interface{}{&i})
}
func (db *Database) Sum(ctx context.Context,  column Column ,qb QB) (value sql.NullInt64, err error) {
	return coreSum(ctx, db,  column, qb)
}
func (tx *Transaction) Sum(ctx context.Context,  column Column ,qb QB) (value sql.NullInt64, err error) {
	return coreSum(ctx, tx, column, qb)
}
func coreSum(ctx context.Context, storager Storager, column Column ,qb QB) (value sql.NullInt64, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SelectRaw = []Raw{{"SUM(" + column.wrapField() + ")", nil}}
	_, err = coreQueryRowScan(ctx, storager, qb, []interface{}{&value}) ; if err != nil {
		return
	}
	return
}

func (db *Database) Update(ctx context.Context, qb QB) (result sql.Result, err error){
	return coreUpdate(ctx, db, qb)
}
func (tx *Transaction) Update(ctx context.Context, qb QB) (result sql.Result, err error){
	return coreUpdate(ctx, tx, qb)
}
func coreUpdate(ctx context.Context, storager Storager, qb QB) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLUpdate()
	query, values := raw.Query, raw.Values
	result, err = storager.getCore().ExecContext(ctx, query, values...)
	if err != nil {return result, err}
	return
}

// func (db *Database) UpdateModel(ctx context.Context, ptr Model, updateData []Update,  qb QB) (result sql.Result, err error){
// 	return coreUpdateModel(ctx, db, ptr, updateData, qb)
// }
// func (tx *Transaction) UpdateModel(ctx context.Context, ptr Model, updateData []Update,  qb QB) (result sql.Result, err error){
// 	return coreUpdateModel(ctx, tx, ptr, updateData, qb)
// }
// func coreUpdateModel(ctx context.Context, storager Storager, ptr Model, updateData []Update,  qb QB) (result sql.Result, err error) {
// 	defer func() { if err != nil { err = xerr.WithStack(err) } }()
// 	rValue := reflect.ValueOf(ptr)
// 	rType := rValue.Type()
// 	if rType.Kind() != reflect.Ptr {
// 		return result, xerr.New("UpdateModel(ctx, ptr) " + rType.String() + " must be ptr")
// 	}
// 	elemValue := rValue.Elem()
// 	elemType := rType.Elem()
// 	for i:=0;i<elemType.NumField();i++ {
// 		fieldType := elemType.Field(i)
// 		fieldValue := elemValue.Field(i)
// 		column, hasDBTag := fieldType.Tag.Lookup("db")
// 		if !hasDBTag {continue}
// 		//  updated time.Time
// 		for _, timeField := range updateTimeField {
// 			if fieldType.Name == timeField {
// 				setTimeNow(fieldValue, fieldType)
// 				// ID IDUser `sq:"ignoreUpdate"`
// 				shouldIgnore := Tag{fieldType.Tag.Get("sq")}.IsIgnoreUpdate()
// 				if !shouldIgnore {
// 					updateData = append(updateData, Update{
// 						Column: Column(column),
// 						Value: fieldValue.Interface(),
// 					})
// 				}
// 			}
// 		}
// 		for dataIndex, data := range updateData {
// 			if len(data.Column) != 0  && column == data.Column.String() {
// 					if data.OnUpdated == nil {
// 						updateData[dataIndex].OnUpdated = func() error {
// 							fieldValue.Set(reflect.ValueOf(data.Value))
// 							return nil
// 						}
// 					}
// 			}
// 		}
// 	}
// 	primaryKey, err := safeGetPrimaryKey(ptr); if err != nil {
// 	    return
// 	}
// 	qb.From = ptr
// 	qb.Update = updateData
// 	qb.Where = primaryKey
// 	qb.SQLChecker = storager.getSQLChecker()
// 	raw := qb.SQLUpdate()
// 	query, values := raw.Query, raw.Values
// 	result, err = storager.getCore().ExecContext(ctx, query, values...)
// 	if err != nil {return result, err}
// 	for _, data := range updateData {
// 		if data.OnUpdated != nil {
// 			updatedErr := data.OnUpdated() ; if updatedErr != nil {
// 				return result, updatedErr
// 			}
// 		}
// 	}
// 	return
// }
func (db *Database) checkIsTestDatabase(ctx context.Context) (err error) {
	var databaseName string
	_, err = db.QueryRowScan(ctx, QB{Raw: Raw{"SELECT DATABASE()", nil}}, []interface{}{&databaseName}) ; if err != nil {
		return
	}
	if strings.HasPrefix(databaseName, "test_") == false {
		return xerr.New("ClearTestData only support delete test database")
	}
	return
}
func (db *Database) ClearTestData(ctx context.Context, qb QB) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	err = db.checkIsTestDatabase(ctx) ; if err != nil {
		return
	}
	return db.HardDelete(ctx, qb)
}
func (db *Database) ClearTestModel(ctx context.Context, model Model, qb QB) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	err = db.checkIsTestDatabase(ctx) ; if err != nil {
		return
	}
	return db.hardDeleteModel(ctx, model, qb)
}
func (db *Database) HardDelete(ctx context.Context, qb QB) (result sql.Result, err error) {
	return coreHardDelete(ctx, db, qb)
}
func (tx *Transaction) HardDelete(ctx context.Context, qb QB) (result sql.Result, err error) {
	return coreHardDelete(ctx, tx, qb)
}
func coreHardDelete(ctx context.Context, storager Storager, qb QB) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLDelete()
	return storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
}
func (db *Database) hardDeleteModel(ctx context.Context, ptr Model, qb QB) (result sql.Result, err error){
	return coreHardDeleteModel(ctx,db, ptr, qb)
}
// func (tx *Transaction) HardDeleteModel(ctx context.Context, ptr Model, qb QB) (result sql.Result, err error){
// 	return coreHardDeleteModel(ctx, tx, ptr, qb)
// }
func coreHardDeleteModel(ctx context.Context, storager Storager, ptr Model, qb QB) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	rValue := reflect.ValueOf(ptr)
	rType := rValue.Type()
	if rType.Kind() != reflect.Ptr {
		return result, xerr.New("UpdateModel(ctx, ptr) " + rType.String() + " must be ptr")
	}
	primaryKey, err := safeGetPrimaryKey(ptr); if err != nil {
		return
	}
	qb.From = ptr
	qb.Where = primaryKey
	qb.Limit = 1

	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLDelete()
	return storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
}
func (db *Database) SoftDelete(ctx context.Context, qb QB) (result sql.Result, err error) {
	return coreSoftDelete(ctx, db, qb)
}
func (tx *Transaction) SoftDelete(ctx context.Context, qb QB) (result sql.Result, err error) {
	return coreSoftDelete(ctx, tx, qb)
}
func coreSoftDelete(ctx context.Context, storager Storager, qb QB) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	softDeleteWhere := qb.From.SoftDeleteWhere()
	if softDeleteWhere.Query == "" {
		err = xerr.New("goclub/sql: SoftDelete(ctx, qb) qb.Form.SoftDeleteWhere().Query can not be empty string" )
		return
	}
	qb.SQLChecker = storager.getSQLChecker()
	softDeleteSet := qb.From.SoftDeleteSet()
	if softDeleteSet.Query == "" {
		err = xerr.New("goclub/sql: SoftDelete()" + qb.From.TableName() + "without soft delete set")
		return
	}
	qb.Update = []Update{
		{Raw: softDeleteSet,},
	}
	raw := qb.SQLUpdate()
	return storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
}
// func (db *Database) SoftDeleteModel(ctx context.Context, ptr Model, qb QB) (result sql.Result, err error){
// 	return coreSoftDeleteModel(ctx, db, ptr, qb)
// }
// func (tx *Transaction) SoftDeleteModel(ctx context.Context, ptr Model, qb QB) (result sql.Result, err error){
// 	return coreSoftDeleteModel(ctx, tx, ptr, qb)
// }
// func coreSoftDeleteModel(ctx context.Context, storager Storager, ptr Model, qb QB) (result sql.Result, err error) {
// 	defer func() { if err != nil { err = xerr.WithStack(err) } }()
// 	rValue := reflect.ValueOf(ptr)
// 	rType := rValue.Type()
// 	if rType.Kind() != reflect.Ptr {
// 		return result, xerr.New("UpdateModel(ctx, ptr) " + rType.String() + " must be ptr")
// 	}
// 	primaryKey, err := safeGetPrimaryKey(ptr); if err != nil {
// 		return
// 	}
// 	qb.From = ptr
// 	qb.Where = primaryKey
// 	qb.Update = []Update{{Raw:ptr.SoftDeleteSet(),}}
// 	qb.Limit = 1
// 	qb.SQLChecker = storager.getSQLChecker()
// 	raw := qb.SQLUpdate()
// 	return storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
// }
func (db *Database) QueryRelation(ctx context.Context, ptr Relation, qb QB) (has bool, err error){
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryRelation(ctx, db, ptr, qb)
}
func (tx *Transaction) QueryRelation(ctx context.Context, ptr Relation, qb QB) (has bool, err error){
	return coreQueryRelation(ctx, tx, ptr, qb)
}
func coreQueryRelation(ctx context.Context, storager Storager, ptr Relation, qb QB) (has bool, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SQLChecker = storager.getSQLChecker()
	qb.Select = TagToColumns(ptr)
	table := table {
		tableName: ptr.TableName(),
		softDeleteWhere: ptr.SoftDeleteWhere,
		// Relation 不需要 update
		softDeleteSet: func() Raw {return Raw{}},
	}
	qb.From = table
	qb.Limit = 1
	qb.Join = ptr.RelationJoin()

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
func (db *Database) QueryRelationSlice(ctx context.Context, relationSlicePtr interface{}, qb QB) (err error) {
	err = qb.mustInTransaction() ; if err != nil {return}
	return coreQueryRelationSlice(ctx, db, relationSlicePtr, qb)
}
func (tx *Transaction) QueryRelationSlice(ctx context.Context, relationSlicePtr interface{}, qb QB) (err error) {
	return coreQueryRelationSlice(ctx, tx, relationSlicePtr, qb)
}
func coreQueryRelationSlice(ctx context.Context, storager Storager, relationSlicePtr interface{}, qb QB) (err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SQLChecker = storager.getSQLChecker()
	ptrType := reflect.TypeOf(relationSlicePtr)
	if ptrType.Kind() != reflect.Ptr {
		return xerr.New("goclub/sql: " + ptrType.String() + "not pointer")
	}
	elemType := ptrType.Elem()
	reflectItemValue := reflect.MakeSlice(elemType, 1,1).Index(0)
	if reflectItemValue.CanAddr() {
		reflectItemValue = reflectItemValue.Addr()
	}
	tablerInterface := reflectItemValue.Interface().(Relation)

	qb.Select = TagToColumns(tablerInterface)
	qb.From = table {
		tableName: tablerInterface.TableName(),
		softDeleteWhere: tablerInterface.SoftDeleteWhere,
		// Relation 不需要 update
		softDeleteSet: func() Raw {return Raw{}},
	}
	qb.Join = tablerInterface.RelationJoin()
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	err = storager.getCore().SelectContext(ctx, relationSlicePtr,query , values...) ; if err != nil {
		return err
	}
	return
}

func (db *Database) Exec(ctx context.Context, query string, values []interface{}) (result sql.Result, err error) {
	return coreExec(ctx, db, query, values)
}
func (tx *Transaction) Exec(ctx context.Context, query string, values []interface{}) (result sql.Result, err error) {
	return coreExec(ctx, tx, query, values)
}
func coreExec(ctx context.Context, storager Storager, query string, values []interface{}) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	return storager.getCore().ExecContext(ctx, query, values... )
}
func (db *Database) ExecQB(ctx context.Context, qb QB, statement Statement) (result sql.Result, err error){
	return coreExecQB(ctx, db, qb, statement)
}
func (tx *Transaction) ExecQB(ctx context.Context, qb QB, statement Statement) (result sql.Result, err error){
	return coreExecQB(ctx, tx, qb, statement)
}
func coreExecQB(ctx context.Context, storager Storager, qb QB, statement Statement) (result sql.Result, err error) {
	defer func() { if err != nil { err = xerr.WithStack(err) } }()
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQL(statement)
	result, err = storager.getCore().ExecContext(ctx, raw.Query, raw.Values...) ; if err != nil {
		return
	}
	return
}
