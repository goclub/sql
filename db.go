package sq

import (
	"context"
	"database/sql"
	xerr "github.com/goclub/error"
	"github.com/jmoiron/sqlx"
	"reflect"
	"strings"
	"time"
)

type Database struct {
	Core              *sqlx.DB
	SQLChecker        SQLChecker
	QueueTimeLocation *time.Location
}

func (db *Database) Ping(ctx context.Context) error {
	return db.Core.PingContext(ctx)
}
func (db *Database) getCore() (core StoragerCore) {
	return db.Core
}
func (db *Database) getSQLChecker() SQLChecker {
	return db.SQLChecker
}
func Open(driverName string, dataSourceName string) (db *Database, dbClose func() error, err error) {
	var coreDatabase *sqlx.DB
	coreDatabase, err = sqlx.Open(driverName, dataSourceName)
	db = &Database{
		Core:              coreDatabase,
		SQLChecker:        &DefaultSQLChecker{},
		QueueTimeLocation: time.Local,
	}
	if err != nil && coreDatabase != nil {
		dbClose = func() error {
			// 忽略 log sync 错误
			_ = Log.Sync()
			return coreDatabase.Close()
		}
	} else {
		dbClose = func() error { return nil }
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
	Log.Warn("Database is nil,maybe you forget sq.Open()")
	return nil
}

var createTimeField = []string{"CreatedAt", "GMTCreate", "CreateTime"}
var updateTimeField = []string{"UpdatedAt", "GMTModified", "UpdateTime"}
var createAndUpdateTimeField = append(createTimeField, updateTimeField...)

func (db *Database) Insert(ctx context.Context, qb QB) (result Result, err error) {
	return coreInsert(ctx, db, qb)
}
func (db *Database) InsertAffected(ctx context.Context, qb QB) (affected int64, err error) {
	return RowsAffected(db.Insert(ctx, qb))
}
func (tx *T) Insert(ctx context.Context, qb QB) (result Result, err error) {
	return coreInsert(ctx, tx, qb)
}
func (tx *T) InsertAffected(ctx context.Context, qb QB) (affected int64, err error) {
	return RowsAffected(tx.Insert(ctx, qb))
}
func coreInsert(ctx context.Context, storager Storager, qb QB) (result Result, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.SQLChecker = storager.getSQLChecker()
	qb.execDebugBefore(ctx, storager, StatementInsert)
	defer qb.execDebugAfter(ctx, storager, StatementInsert)
	return coreExecQB(ctx, storager, qb, StatementInsert)
}

func (db *Database) InsertModel(ctx context.Context, ptr Model, qb QB) (err error) {
	if _, err = coreInsertModel(ctx, db, ptr, qb); err != nil {
		return
	}
	return
}
func (db *Database) InsertModelAffected(ctx context.Context, ptr Model, qb QB) (affected int64, err error) {
	return RowsAffected(coreInsertModel(ctx, db, ptr, qb))
}
func (tx *T) InsertModel(ctx context.Context, ptr Model, qb QB) (err error) {
	if _, err = coreInsertModel(ctx, tx, ptr, qb); err != nil {
		return
	}
	return
}
func (tx *T) InsertModelAffected(ctx context.Context, ptr Model, qb QB) (affected int64, err error) {
	return RowsAffected(coreInsertModel(ctx, tx, ptr, qb))
}

func coreInsertModel(ctx context.Context, storager Storager, ptr Model, qb QB) (result Result, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	err = ptr.BeforeInsert()
	if err != nil {
		return
	}
	if qb.From != nil {
		Log.Warn("InsertModel(ctx, qb, model) qb.From need be nil")
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
	qb.execDebugBefore(ctx, storager, StatementInsert)
	defer qb.execDebugAfter(ctx, storager, StatementInsert)
	result.core, err = storager.getCore().ExecContext(ctx, query, values...)
	if err != nil {
		return
	}
	err = ptr.AfterInsert(result)
	if err != nil {
		return
	}
	return
}
func insertEachField(elemValue reflect.Value, elemType reflect.Type, handle func(column string, fieldType reflect.StructField, fieldValue reflect.Value)) {
	for i := 0; i < elemType.NumField(); i++ {
		fieldType := elemType.Field(i)
		fieldValue := elemValue.Field(i)
		// `db:"name"`
		column, hasDBTag := fieldType.Tag.Lookup("db")
		if fieldType.Anonymous == true {
			insertEachField(fieldValue, fieldValue.Type(), handle)
			continue
		}
		if !hasDBTag {
			continue
		}
		if column == "" {
			continue
		}
		// `sq:"ignoreInsert"`
		shouldIgnoreInsert := Tag{fieldType.Tag.Get("sq")}.IsIgnoreInsert()
		if shouldIgnoreInsert {
			continue
		}
		// created updated time.Time
		for _, timeField := range createAndUpdateTimeField {
			if fieldType.Name == timeField {
				setTimeNow(fieldValue, fieldType)
			}
		}
		handle(column, fieldType, fieldValue)
	}
}
func (db *Database) QueryRow(ctx context.Context, qb QB, desc []interface{}) (has bool, err error) {
	err = qb.mustInTransaction()
	if err != nil {
		return
	}
	return coreQueryRowScan(ctx, db, qb, desc)
}
func (tx *T) QueryRow(ctx context.Context, qb QB, desc []interface{}) (has bool, err error) {
	return coreQueryRowScan(ctx, tx, qb, desc)
}
func coreQueryRowScan(ctx context.Context, storager Storager, qb QB, desc []interface{}) (has bool, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.SQLChecker = storager.getSQLChecker()
	qb.Limit = 1
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	row := storager.getCore().QueryRowxContext(ctx, query, values...)
	scanErr := row.Scan(desc...)
	has, err = CheckRowScanErr(scanErr)
	if err != nil {
		return
	}
	return
}
func (db *Database) QuerySliceScaner(ctx context.Context, qb QB, scan Scaner) (err error) {
	err = qb.mustInTransaction()
	if err != nil {
		return
	}
	return coreQuerySliceScaner(ctx, db, qb, scan)
}
func (tx *T) QuerySliceScaner(ctx context.Context, qb QB, scan Scaner) error {
	return coreQuerySliceScaner(ctx, tx, qb, scan)
}
func coreQuerySliceScaner(ctx context.Context, storager Storager, qb QB, scan Scaner) (err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	rows, err := storager.getCore().QueryxContext(ctx, query, values...)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			return
		}
	}()
	for rows.Next() {
		err := scan(rows)
		if err != nil {
			return err
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return rowsErr
	}
	return nil
}
func (db *Database) Query(ctx context.Context, ptr Tabler, qb QB) (has bool, err error) {
	err = qb.mustInTransaction()
	if err != nil {
		return
	}
	return coreQuery(ctx, db, ptr, qb)
}
func (tx *T) Query(ctx context.Context, ptr Tabler, qb QB) (has bool, err error) {
	return coreQuery(ctx, tx, ptr, qb)
}

func coreQuery(ctx context.Context, storager Storager, ptr Tabler, qb QB) (has bool, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.SQLChecker = storager.getSQLChecker()
	qb.Limit = 1
	if qb.From == nil {
		qb.From = ptr
	}
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	row := storager.getCore().QueryRowxContext(ctx, query, values...)
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	scanErr := row.StructScan(ptr)
	has, err = CheckRowScanErr(scanErr)
	if err != nil {
		return
	}
	return
}

func (db *Database) QuerySlice(ctx context.Context, slicePtr interface{}, qb QB) (err error) {
	err = qb.mustInTransaction()
	if err != nil {
		return
	}
	return coreQuerySlice(ctx, db, slicePtr, qb)
}
func (tx *T) QuerySlice(ctx context.Context, slicePtr interface{}, qb QB) (err error) {
	return coreQuerySlice(ctx, tx, slicePtr, qb)
}
func coreQuerySlice(ctx context.Context, storager Storager, slicePtr interface{}, qb QB) (err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.SQLChecker = storager.getSQLChecker()
	ptrType := reflect.TypeOf(slicePtr)
	if ptrType.Kind() != reflect.Ptr {
		return xerr.New("goclub/sql: " + ptrType.String() + "not pointer")
	}
	if qb.From == nil && qb.FromRaw.TableName.Query == "" && qb.Raw.Query == "" {
		elemType := ptrType.Elem()
		reflectItemValue := reflect.MakeSlice(elemType, 1, 1).Index(0)
		if reflectItemValue.CanAddr() {
			reflectItemValue = reflectItemValue.Addr()
		}
		tablerInterface := reflectItemValue.Interface().(Tabler)
		qb.From = tablerInterface
	}
	if qb.From == nil && qb.FromRaw.TableName.Query == "" && qb.Raw.Query == "" {
		// 如果设置了 qb.Form 但没有设置 qb.Select 可能会导致 select * ,这种情况在代码已经在线上运行时但是表变动了时会很危险
		if len(qb.Select) == 0 && len(qb.SelectRaw) == 0 {
			err = xerr.New("goclub/sql: QuerySlice(ctx, slice, qb) if qb.Form/qb.FromRaw/qb.Raw not zero value, then qb.Select or qb.SelectRaw can not be nil, or you can set qb.Form/qb.FromRaw/ be nil")
			return
		}
	}
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	return storager.getCore().SelectContext(ctx, slicePtr, query, values...)
}
func (db *Database) Count(ctx context.Context, from Tabler, qb QB) (count uint64, err error) {
	err = qb.mustInTransaction()
	if err != nil {
		return
	}
	return coreCount(ctx, db, from, qb)
}
func (tx *T) Count(ctx context.Context, from Tabler, qb QB) (count uint64, err error) {
	return coreCount(ctx, tx, from, qb)
}
func coreCount(ctx context.Context, storager Storager, from Tabler, qb QB) (count uint64, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.From = from
	qb.SQLChecker = storager.getSQLChecker()
	if len(qb.SelectRaw) == 0 {
		qb.SelectRaw = []Raw{{"COUNT(*)", nil}}
	}
	qb.limitRaw = limitRaw{Valid: true, Limit: 0}
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	var has bool
	has, err = coreQueryRowScan(ctx, storager, qb, []interface{}{&count})
	if err != nil {
		return
	}
	if has == false {
		raw := qb.SQLSelect()
		query := raw.Query
		return 0, xerr.New("goclub/sql: Count() " + query + "not found data")
	}
	return
}

// if you need query data exited SELECT "has" FROM user WHERE id = ? better than SELECT count(*) FROM user where id = ?
func (db *Database) Has(ctx context.Context, from Tabler, qb QB) (has bool, err error) {
	err = qb.mustInTransaction()
	if err != nil {
		return
	}
	return coreHas(ctx, db, from, qb)
}
func (tx *T) Has(ctx context.Context, from Tabler, qb QB) (has bool, err error) {
	return coreHas(ctx, tx, from, qb)
}
func coreHas(ctx context.Context, storager Storager, from Tabler, qb QB) (has bool, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.From = from
	qb.SQLChecker = storager.getSQLChecker()
	qb.SelectRaw = []Raw{{`1`, nil}}
	qb.Limit = 1
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	var i int
	return coreQueryRowScan(ctx, storager, qb, []interface{}{&i})
}
func (db *Database) SumInt64(ctx context.Context, from Tabler, column Column, qb QB) (value sql.NullInt64, err error) {
	err = coreSum(ctx, db, from, column, qb, &value)
	if err != nil {
		return
	}
	return value, err
}
func (tx *T) SumInt64(ctx context.Context, from Tabler, column Column, qb QB) (value sql.NullInt64, err error) {
	err = coreSum(ctx, tx, from, column, qb, &value)
	if err != nil {
		return
	}
	return value, err
}
func (db *Database) SumFloat64(ctx context.Context, from Tabler, column Column, qb QB) (value sql.NullFloat64, err error) {
	err = coreSum(ctx, db, from, column, qb, &value)
	if err != nil {
		return
	}
	return value, err
}
func (tx *T) SumFloat64(ctx context.Context, from Tabler, column Column, qb QB) (value sql.NullFloat64, err error) {
	err = coreSum(ctx, tx, from, column, qb, &value)
	if err != nil {
		return
	}
	return value, err
}
func coreSum(ctx context.Context, storager Storager, from Tabler, column Column, qb QB, valuePtr interface{}) (err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.From = from
	qb.SQLChecker = storager.getSQLChecker()
	qb.SelectRaw = []Raw{{"SUM(" + column.wrapField() + ")", nil}}
	qb.limitRaw.Valid = true
	qb.limitRaw.Limit = 0
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	_, err = coreQueryRowScan(ctx, storager, qb, []interface{}{valuePtr})
	if err != nil {
		return
	}
	return
}

func (db *Database) Update(ctx context.Context, from Tabler, qb QB) (err error) {
	if _, err = coreUpdate(ctx, db, from, qb); err != nil {
		return
	}
	return
}
func (db *Database) UpdateAffected(ctx context.Context, from Tabler, qb QB) (affected int64, err error) {
	return RowsAffected(coreUpdate(ctx, db, from, qb))
}
func (tx *T) Update(ctx context.Context, from Tabler, qb QB) (err error) {
	if _, err = coreUpdate(ctx, tx, from, qb); err != nil {
		return
	}
	return
}
func (tx *T) UpdateAffected(ctx context.Context, from Tabler, qb QB) (affected int64, err error) {
	return RowsAffected(coreUpdate(ctx, tx, from, qb))
}
func coreUpdate(ctx context.Context, storager Storager, from Tabler, qb QB) (result Result, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	if qb.From == nil {
		qb.From = from
	}
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLUpdate()
	query, values := raw.Query, raw.Values
	result.core, err = storager.getCore().ExecContext(ctx, query, values...)
	qb.execDebugBefore(ctx, storager, StatementUpdate)
	defer qb.execDebugAfter(ctx, storager, StatementUpdate)
	if err != nil {
		return result, err
	}
	return
}

func (db *Database) checkIsTestDatabase(ctx context.Context) (err error) {
	var databaseName string
	_, err = db.QueryRow(ctx, QB{Raw: Raw{"SELECT DATABASE()", nil}}, []interface{}{&databaseName})
	if err != nil {
		return
	}
	if strings.HasPrefix(databaseName, "test_") == false {
		return xerr.New("ClearTestData only support delete test database")
	}
	return
}
func (db *Database) ClearTestData(ctx context.Context, from Tabler, qb QB) (err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	err = db.checkIsTestDatabase(ctx)
	if err != nil {
		return
	}
	return db.HardDelete(ctx, from, qb)
}
func (db *Database) HardDelete(ctx context.Context, from Tabler, qb QB) (err error) {
	if _, err = coreHardDelete(ctx, db, from, qb); err != nil {
		return
	}
	return
}
func (db *Database) HardDeleteAffected(ctx context.Context, from Tabler, qb QB) (affected int64, err error) {
	return RowsAffected(coreHardDelete(ctx, db, from, qb))
}
func (tx *T) HardDelete(ctx context.Context, from Tabler, qb QB) (err error) {
	if _, err = coreHardDelete(ctx, tx, from, qb); err != nil {
		return
	}
	return
}
func (tx *T) HardDeleteAffected(ctx context.Context, from Tabler, qb QB) (affected int64, err error) {
	return RowsAffected(coreHardDelete(ctx, tx, from, qb))
}
func coreHardDelete(ctx context.Context, storager Storager, from Tabler, qb QB) (result Result, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	if qb.From == nil {
		qb.From = from
	}
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQLDelete()
	result.core, err = storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
	if err != nil {
		return
	}
	return
}

// func (db *Database) hardDeleteModel(ctx context.Context, ptr Model, qb QB) (result Result, err error){
// 	return coreHardDeleteModel(ctx,db, ptr, qb)
// }
// func (tx *T) HardDeleteModel(ctx context.Context, ptr Model, qb QB) (result Result, err error){
// 	return coreHardDeleteModel(ctx, tx, ptr, qb)
// }
// func coreHardDeleteModel(ctx context.Context, storager Storager, ptr Model, qb QB) (result Result, err error) {
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
// 	qb.Limit = 1
//
// 	qb.SQLChecker = storager.getSQLChecker()
// 	raw := qb.SQLDelete()
// 	qb.execDebugBefore(ctx, storager, StatementUpdate)
// 	defer qb.execDebugAfter(ctx, storager, StatementUpdate)
// 	return storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
// }
func (db *Database) SoftDelete(ctx context.Context, from Tabler, qb QB) (err error) {
	if _, err = coreSoftDelete(ctx, db, from, qb); err != nil {
		return
	}
	return
}
func (db *Database) SoftDeleteAffected(ctx context.Context, from Tabler, qb QB) (affected int64, err error) {
	return RowsAffected(coreSoftDelete(ctx, db, from, qb))
}

func (tx *T) SoftDelete(ctx context.Context, from Tabler, qb QB) (err error) {
	if _, err = coreSoftDelete(ctx, tx, from, qb); err != nil {
		return
	}
	return
}
func (tx *T) SoftDeleteAffected(ctx context.Context, from Tabler, qb QB) (affected int64, err error) {
	return RowsAffected(coreSoftDelete(ctx, tx, from, qb))
}
func coreSoftDelete(ctx context.Context, storager Storager, from Tabler, qb QB) (result Result, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	if qb.From == nil {
		qb.From = from
	}
	softDeleteWhere := qb.From.SoftDeleteWhere()
	if softDeleteWhere.Query == "" {
		err = xerr.New("goclub/sql: SoftDelete(ctx, qb) qb.Form.SoftDeleteWhere().Query can not be empty string")
		return
	}
	qb.SQLChecker = storager.getSQLChecker()
	softDeleteSet := qb.From.SoftDeleteSet()
	if softDeleteSet.Query == "" {
		err = xerr.New("goclub/sql: SoftDelete()" + qb.From.TableName() + "without soft delete set")
		return
	}
	qb.Set = []Update{
		{Raw: softDeleteSet},
	}
	raw := qb.SQLUpdate()
	qb.execDebugBefore(ctx, storager, StatementUpdate)
	defer qb.execDebugAfter(ctx, storager, StatementUpdate)
	result.core, err = storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
	if err != nil {
		return
	}
	return
}
func (db *Database) QueryRelation(ctx context.Context, ptr Relation, qb QB) (has bool, err error) {
	err = qb.mustInTransaction()
	if err != nil {
		return
	}
	return coreQueryRelation(ctx, db, ptr, qb)
}
func (tx *T) QueryRelation(ctx context.Context, ptr Relation, qb QB) (has bool, err error) {
	return coreQueryRelation(ctx, tx, ptr, qb)
}
func coreQueryRelation(ctx context.Context, storager Storager, ptr Relation, qb QB) (has bool, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.SQLChecker = storager.getSQLChecker()
	qb.Select = TagToColumns(ptr)
	table := table{
		tableName:       ptr.TableName(),
		softDeleteWhere: ptr.SoftDeleteWhere,
		// Relation 不需要 update
		softDeleteSet: func() Raw { return Raw{} },
	}
	qb.From = table
	qb.Limit = 1
	qb.Join = ptr.RelationJoin()

	qb.SQLChecker = storager.getSQLChecker()
	qb.Limit = 1

	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	row := storager.getCore().QueryRowxContext(ctx, query, values...)
	scanErr := row.StructScan(ptr)
	has, err = CheckRowScanErr(scanErr)
	if err != nil {
		return
	}
	return
}
func (db *Database) QueryRelationSlice(ctx context.Context, relationSlicePtr interface{}, qb QB) (err error) {
	err = qb.mustInTransaction()
	if err != nil {
		return
	}
	return coreQueryRelationSlice(ctx, db, relationSlicePtr, qb)
}
func (tx *T) QueryRelationSlice(ctx context.Context, relationSlicePtr interface{}, qb QB) (err error) {
	return coreQueryRelationSlice(ctx, tx, relationSlicePtr, qb)
}
func coreQueryRelationSlice(ctx context.Context, storager Storager, relationSlicePtr interface{}, qb QB) (err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.SQLChecker = storager.getSQLChecker()
	ptrType := reflect.TypeOf(relationSlicePtr)
	if ptrType.Kind() != reflect.Ptr {
		return xerr.New("goclub/sql: " + ptrType.String() + "not pointer")
	}
	elemType := ptrType.Elem()
	reflectItemValue := reflect.MakeSlice(elemType, 1, 1).Index(0)
	if reflectItemValue.CanAddr() {
		reflectItemValue = reflectItemValue.Addr()
	}
	tablerInterface := reflectItemValue.Interface().(Relation)

	qb.Select = TagToColumns(tablerInterface)
	qb.From = table{
		tableName:       tablerInterface.TableName(),
		softDeleteWhere: tablerInterface.SoftDeleteWhere,
		// Relation 不需要 update
		softDeleteSet: func() Raw { return Raw{} },
	}
	qb.Join = tablerInterface.RelationJoin()
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
	qb.execDebugBefore(ctx, storager, StatementSelect)
	defer qb.execDebugAfter(ctx, storager, StatementSelect)
	err = storager.getCore().SelectContext(ctx, relationSlicePtr, query, values...)
	if err != nil {
		return err
	}
	return
}

func (db *Database) Exec(ctx context.Context, query string, values []interface{}) (result Result, err error) {
	return coreExec(ctx, db, query, values)
}
func (tx *T) Exec(ctx context.Context, query string, values []interface{}) (result Result, err error) {
	return coreExec(ctx, tx, query, values)
}
func coreExec(ctx context.Context, storager Storager, query string, values []interface{}) (result Result, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	result.core, err = storager.getCore().ExecContext(ctx, query, values...)
	if err != nil {
		return
	}
	return
}
func (db *Database) ExecQB(ctx context.Context, qb QB, statement Statement) (result Result, err error) {
	return coreExecQB(ctx, db, qb, statement)
}
func (db *Database) ExecQBAffected(ctx context.Context, qb QB, statement Statement) (affected int64, err error) {
	return RowsAffected(coreExecQB(ctx, db, qb, statement))
}
func (tx *T) ExecQB(ctx context.Context, qb QB, statement Statement) (result Result, err error) {
	return coreExecQB(ctx, tx, qb, statement)
}
func (tx *T) ExecQBAffected(ctx context.Context, qb QB, statement Statement) (affected int64, err error) {
	return RowsAffected(coreExecQB(ctx, tx, qb, statement))
}
func coreExecQB(ctx context.Context, storager Storager, qb QB, statement Statement) (result Result, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	qb.SQLChecker = storager.getSQLChecker()
	raw := qb.SQL(statement)
	result.core, err = storager.getCore().ExecContext(ctx, raw.Query, raw.Values...)
	if err != nil {
		return
	}
	return
}
func (db *Database) LastQueryCost(ctx context.Context) (lastQueryCost float64, err error) {
	return coreLastQueryCost(ctx, db)
}
func (tx *T) LastQueryCost(ctx context.Context) (lastQueryCost float64, err error) {
	return coreLastQueryCost(ctx, tx)
}
func coreLastQueryCost(ctx context.Context, storager Storager) (lastQueryCost float64, err error) {
	defer func() {
		if err != nil {
			err = xerr.WithStack(err)
		}
	}()
	rows := storager.getCore().QueryRowxContext(ctx, `show status like 'last_query_cost'`)
	if err != nil {
		return
	}
	var name string
	err = rows.Scan(&name, &lastQueryCost)
	if err != nil {
		return
	}
	return
}
func (db *Database) PrintLastQueryCost(ctx context.Context) {
	corePrintLastQueryCost(ctx, db)
}
func (tx *T) PrintLastQueryCost(ctx context.Context) {
	corePrintLastQueryCost(ctx, tx)
}
func corePrintLastQueryCost(ctx context.Context, storager Storager) {
	cost, err := coreLastQueryCost(ctx, storager)
	if err != nil {
		Log.Debug("error", "error", err)
	}
	Log.Debug("last_query_cost", "cost", cost)
}
