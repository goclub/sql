package sq

import (
	"context"
	"crypto/rand"
	xerr "github.com/goclub/error"
	"math/big"
	"sort"
	"strings"
	"testing"
	"time"
)

// sq.Set(column, value)
type Update struct {
	Column Column
	Value  interface{}
	Raw    Raw
}
type updates []Update

func OnlyUseInTestToUpdates(t *testing.T, list []Update) updates {
	return list
}
func (u updates) Set(column Column, value interface{}) updates {
	if op, ok := value.(OP); ok {
		value = op.Values[0]
		Log.Warn(`sq.Set(` + column.String() + `, value) value can not be sq.Equal(v) or sq.OP{}, may be you need use like sq.Set("` + column.String() + `", v)`)
	}
	u = append(u, Update{
		Column: column,
		Value:  value,
	})
	return u
}
func (u updates) SetRaw(query string, values ...interface{}) updates {
	for i, value := range values {
		if op, ok := value.(OP); ok {
			values[i] = op.Values[0]
			Log.Warn("sq.SetRaw(query, values) values element can not be sq.Equal(v) or sq.OP{}, may be you need use like sq.Set(\"id\", taskID)")
		}
	}
	u = append(u, Update{
		Raw: Raw{query, values},
	})
	return u
}
func Set(column Column, value interface{}) updates {
	return updates{}.Set(column, value)
}
func SetMap(data map[Column]interface{}) updates {
	var list []Update
	for column, value := range data {
		list = append(list, Update{
			Column: column,
			Value:  value,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return strings.Compare(list[i].Column.String(), list[j].Column.String()) == -1
	})
	return list
}
func SetRaw(query string, value ...interface{}) updates {
	return updates{}.SetRaw(query, value...)
}

type InsertMultiple struct {
	Column []Column
	Values [][]interface{}
}
type Values []Insert

type Insert struct {
	Column Column
	Value  interface{}
}

type QB struct {
	Select    []Column
	SelectRaw []Raw

	From    Tabler
	from    string
	FromRaw FromRaw

	DisableSoftDelete bool
	softDelete        Raw

	UnionTable UnionTable

	Index string

	Set []Update
	// UPDATE IGNORE
	UseUpdateIgnore bool
	Insert          Values
	InsertMultiple  InsertMultiple
	// INSERT IGNORE INTO
	UseInsertIgnoreInto bool

	Where           []Condition
	WhereOR         [][]Condition
	WhereRaw        Raw
	WhereAllowEmpty bool

	OrderBy    []OrderBy
	OrderByRaw Raw

	GroupBy    []Column
	GroupByRaw Raw

	Having    []Condition
	HavingRaw Raw

	Limit    uint64
	limitRaw limitRaw
	Offset   uint64

	Lock SelectLock

	Join []Join
	Raw  Raw

	Debug           bool
	debugData       struct{ id uint64 }
	PrintSQL        bool
	Explain         bool
	RunTime         bool
	elapsedTimeData struct {
		startTime time.Time
	}
	LastQueryCost bool

	Review            string
	Reviews           []string
	SQLChecker        SQLChecker
	disableSQLChecker bool
}

func (qb QB) mustInTransaction() error {
	if len(qb.Lock) == 0 {
		return nil
	}
	return xerr.New("goclub/sql: SELECT " + qb.Lock.String() + " must exec in transaction")
}

type FromRaw struct {
	TableName       Raw
	SoftDeleteWhere Raw
}
type OrderBy struct {
	Column Column
	Type   orderByType
}
type orderByType uint8

const (
	// 默认降序
	ASC orderByType = iota
	DESC
)

type SelectLock string

func (s SelectLock) String() string {
	return string(s)
}

const FORSHARE SelectLock = "FOR SHARE"
const FORUPDATE SelectLock = "FOR UPDATE"

type UnionTable struct {
	Tables   []QB
	UnionAll bool
}

func (union UnionTable) SQLSelect() (raw Raw) {
	var sqlList stringQueue
	var subQueryList []string
	for _, table := range union.Tables {
		subQV := table.SQLSelect()
		subQueryList = append(subQueryList, "("+subQV.Query+")")
		raw.Values = append(raw.Values, subQV.Values...)
	}
	unionText := "UNION"
	if union.UnionAll {
		unionText += " ALL"
	}
	sqlList.Push(strings.Join(subQueryList, " "+unionText+" "))
	raw.Query = sqlList.Join(" ")
	return
}

type limitRaw struct {
	Valid bool
	Limit uint64
}
type JoinType string

func (t JoinType) String() string {
	return string(t)
}

const InnerJoin JoinType = "INNER JOIN"
const LeftJoin JoinType = "LEFT JOIN"
const RightJoin JoinType = "RIGHT JOIN"
const FullOuterJoin JoinType = "FULL OUTER JOIN"
const CrossJoin JoinType = "CROSS JOIN"

type Join struct {
	Type      JoinType
	TableName string
	On        string
}
type Column string

func (c Column) String() string { return string(c) }
func (c Column) wrapField() string {
	s := c.String()
	return "`" + strings.ReplaceAll(s, ".", "`.`") + "`"
}
func (c Column) wrapFieldWithAS() string {
	s := c.String()
	column := c.wrapField()
	if strings.Contains(s, ".") {
		column += ` AS '` + s + `'`
	}
	return column
}

type Statement string

const StatementSelect Statement = "SELECT"
const StatementUpdate Statement = "UPDATE"
const StatementDelete Statement = "DELETE"
const StatementInsert Statement = "INSERT"

func (s Statement) String() string { return string(s) }

func (qb QB) SQL(statement Statement) Raw {
	originQB := qb
	if len(qb.Raw.Query) != 0 {
		return qb.Raw
	}
	if statement != StatementInsert && qb.whereIsEmpty() && qb.WhereAllowEmpty == false {
		cloneQB := originQB
		cloneQB.WhereAllowEmpty = true

		warning := "query:" + "\n" +
			"\t" + cloneQB.SQL(statement).Query + "\n" +
			"If you need where is empty, set qb.WhereAllowEmpty = true"
		Log.Warn("Maybe you forget qb.Where\n" + warning)
	}
	var values []interface{}
	var sqlList stringQueue
	if statement == StatementSelect && qb.UnionTable.Tables != nil {
		unionRaw := qb.UnionTable.SQLSelect()
		sqlList.Push(unionRaw.Query)
		values = append(values, unionRaw.Values...)
	}
	if qb.From != nil {
		qb.from = "`" + qb.From.TableName() + "`"
		switch statement {
		case StatementSelect,
			StatementUpdate:
			qb.softDelete = qb.From.SoftDeleteWhere()
		case StatementInsert:
		case StatementDelete:
		default:
			panic(xerr.New("statement can not be " + statement.String()))
		}
	}
	if qb.FromRaw.TableName.Query != "" {
		qb.from = qb.FromRaw.TableName.Query
		values = append(values, qb.FromRaw.TableName.Values...)
		qb.softDelete = qb.FromRaw.SoftDeleteWhere
	}
	switch statement {
	case StatementSelect:
		if qb.UnionTable.Tables == nil {
			sqlList.Push("SELECT")
			if qb.SelectRaw == nil {
				inputSelectLen := len(qb.Select)
				if qb.From != nil && inputSelectLen == 0 {
					qb.Select = TagToColumns(qb.From)
				}
				newSelectLen := len(qb.Select)
				if newSelectLen == 0 {
					warningTitle := "goclub/sql: (NO SELECT FIELD)"
					var warning string
					if qb.From != nil {
						warning = "qb.From field does not have db struct tag or you forget set qb.Select"
					} else {
						warning = "qb.Select is empty and qb.Form is nil, maybe you forget set qb.Select"
					}
					Log.Warn(warningTitle + "\n" + warning)
					return Raw{Query: warningTitle + " " + warning}
				} else {

				}
				sqlList.Push(strings.Join(columnsToStringsWithAS(qb.Select), ", "))
			} else {
				var rawColumns []string
				for _, raws := range qb.SelectRaw {
					rawColumns = append(rawColumns, raws.Query)
					values = append(values, raws.Values...)
				}
				sqlList.Push(strings.Join(rawColumns, ", "))
			}
			sqlList.Push("FROM")
			sqlList.Push(qb.from)
		}
		if qb.Index != "" {
			sqlList.Push(qb.Index)
		}
		for _, join := range qb.Join {
			sqlList.Push(join.Type.String())
			sqlList.Push("`" + join.TableName + "`")
			sqlList.Push("ON")
			sqlList.Push(join.On)
		}
	case StatementUpdate:
		sqlList.Push("UPDATE")
		if qb.UseUpdateIgnore {
			sqlList.Push("IGNORE")
		}
		sqlList.Push(qb.from)
		sqlList.Push("SET")
		var sets []string
		for _, data := range qb.Set {
			if len(data.Raw.Query) != 0 {
				sets = append(sets, data.Raw.Query)
				values = append(values, data.Raw.Values...)
			} else {
				sets = append(sets, data.Column.wrapField()+" = ?")
				values = append(values, data.Value)
			}
		}
		sqlList.Push(strings.Join(sets, ", "))
	case StatementDelete:
		sqlList.Push("DELETE FROM")
		sqlList.Push(qb.from)
	case StatementInsert:
		if qb.UseInsertIgnoreInto {
			sqlList.Push("INSERT IGNORE INTO")
		} else {
			sqlList.Push("INSERT INTO")
		}

		sqlList.Push(qb.from)
		if len(qb.Insert) != 0 {
			var insertValues []interface{}
			for _, insert := range qb.Insert {
				qb.InsertMultiple.Column = append(qb.InsertMultiple.Column, insert.Column)
				insertValues = append(insertValues, insert.Value)
			}
			qb.InsertMultiple.Values = append(qb.InsertMultiple.Values, insertValues)
		}

		var columns []string
		for _, column := range qb.InsertMultiple.Column {
			columns = append(columns, column.wrapField())
		}
		var allPlaceholders []string
		for _, value := range qb.InsertMultiple.Values {
			var rowPlaceholder []string
			for _, v := range value {
				var insertRaw Raw
				var hasInsertRaw bool
				switch item := v.(type) {
				case sqlInsertRawer:
					insertRaw.Query, insertRaw.Values = item.SQLInsertRaw()
					hasInsertRaw = true
				}
				if hasInsertRaw {
					rowPlaceholder = append(rowPlaceholder, insertRaw.Query)
					values = append(values, insertRaw.Values...)
				} else {
					rowPlaceholder = append(rowPlaceholder, "?")
					values = append(values, v)
				}
			}
			allPlaceholders = append(allPlaceholders, "("+strings.Join(rowPlaceholder, ",")+")")

		}
		sqlList.Push("(" + strings.Join(columns, ",") + ")")
		sqlList.Push("VALUES")
		sqlList.Push(strings.Join(allPlaceholders, ","))
	default:
		panic(xerr.New("incorrect statement"))
	}
	// where
	{
		var whereString string
		var whereRaw Raw
		if qb.WhereRaw.Query != "" {
			whereRaw = qb.WhereRaw
		} else {
			tooMuchWhere := len(qb.Where) != 0 && len(qb.WhereOR) != 0
			if tooMuchWhere {
				panic(xerr.New("if qb.WhereOR not empty, then qb.Where must be nil"))
			}
			if len(qb.Where) != 0 && len(qb.WhereOR) == 0 {
				qb.WhereOR = append(qb.WhereOR, qb.Where)
			}
			whereRaw = ConditionsSQL(qb.WhereOR)
		}
		var whereValues []interface{}
		whereString, whereValues = whereRaw.Query, whereRaw.Values
		values = append(values, whereValues...)
		var disableWhereIsEmpty bool
		if statement == StatementDelete || statement == StatementUpdate {
			disableWhereIsEmpty = true
		}
		if disableWhereIsEmpty && len(strings.TrimSpace(whereString)) == 0 {
			return Raw{"goclub/sql:(MAYBE_FORGET_WHERE)", nil}
		}
		if !qb.DisableSoftDelete {
			needSoftDelete := qb.softDelete.Query != ""
			if needSoftDelete {
				whereSoftDelete := qb.softDelete
				values = append(values, whereSoftDelete.Values...)
				if len(whereString) != 0 {
					whereString += " AND " + whereSoftDelete.Query
				} else {
					whereString += whereSoftDelete.Query
				}
			}
		}
		if len(whereString) != 0 {
			sqlList.Push("WHERE")
			sqlList.Push(whereString)
		}
	}
	// group by
	if qb.GroupByRaw.Query != "" {
		sqlList.Push("GROUP BY")
		sqlList.Push(qb.GroupByRaw.Query)
		values = append(values, qb.GroupByRaw.Values...)
	} else if len(qb.GroupBy) != 0 {
		sqlList.Push("GROUP BY")
		sqlList.Push(strings.Join(columnsToStrings(qb.GroupBy), ", "))
	}
	// having
	if qb.HavingRaw.Query != "" {
		sqlList.Push("HAVING")
		sqlList.Push(qb.HavingRaw.Query)
		values = append(values, qb.HavingRaw.Values...)
	} else if len(qb.Having) != 0 {
		sqlList.Push("HAVING")
		havaingRaw := ConditionsSQL([][]Condition{qb.Having})
		sqlList.Push(havaingRaw.Query)
		values = append(values, havaingRaw.Values...)
	}
	// order by
	if qb.OrderByRaw.Query != "" {
		sqlList.Push("ORDER BY")
		sqlList.Push(qb.OrderByRaw.Query)
		values = append(values, qb.OrderByRaw.Values...)
	} else if len(qb.OrderBy) != 0 {
		sqlList.Push("ORDER BY")
		var orderList stringQueue
		for _, order := range qb.OrderBy {
			switch order.Type {
			case ASC:
				orderList.Push(order.Column.wrapField() + " ASC")
			case DESC:
				orderList.Push(order.Column.wrapField() + " DESC")
			}
		}
		sqlList.Push(orderList.Join(", "))
	}
	// limit
	limit := qb.Limit
	// 优先使用 qb.limitRaw, 因为 db.Count 需要用到
	if qb.limitRaw.Valid {
		limit = qb.limitRaw.Limit
	}
	if limit != 0 {
		sqlList.Push("LIMIT ?")
		values = append(values, qb.Limit)
	}
	// offset
	if qb.Offset != 0 {
		sqlList.Push("OFFSET ?")
		values = append(values, qb.Offset)
	}
	// lock
	if len(qb.Lock) != 0 {
		sqlList.Push(qb.Lock.String())
	}
	query := sqlList.Join(" ")
	defer func() {
		if qb.Review != "" {
			qb.Reviews = append(qb.Reviews, qb.Review)
		}
		if len(qb.Reviews) != 0 && qb.disableSQLChecker == false {
			matched, refs, err := qb.SQLChecker.Check(qb.Reviews, query)
			if err != nil {
				qb.SQLChecker.TrackFail(qb.debugData.id, err, qb.Reviews, query, "")
			} else {
				if matched == false {
					qb.SQLChecker.TrackFail(qb.debugData.id, err, qb.Reviews, query, refs)
				}
			}
		}
	}()
	return Raw{query, values}
}
func (qb QB) SQLSelect() Raw {
	return qb.SQL(StatementSelect)
}
func (qb QB) SQLInsert() Raw {
	return qb.SQL(StatementInsert)
}
func (qb QB) SQLUpdate() Raw {
	return qb.SQL(StatementUpdate)
}
func (qb QB) SQLDelete() Raw {
	return qb.SQL(StatementDelete)
}
func (qb QB) Paging(page uint64, perPage uint32) QB {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 10
	}
	qb.Offset = (page - 1) * uint64(perPage)
	qb.Limit = uint64(perPage)
	return qb
}
func (qb *QB) execDebugBefore(ctx context.Context, storager Storager, statement Statement) {
	var err error
	defer func() {
		if err != nil {
			Log.Debug("error", "error", err)
		}
	}()
	debugID, err := rand.Int(rand.Reader, new(big.Int).SetUint64(9999))
	if err != nil {
		// 这个错误故意不处理
		Log.Debug("error", "error", err)
		err = nil
	}
	qb.debugData.id = debugID.Uint64()
	var debugLog []string
	if qb.Debug {
		debugLog = append(debugLog, "Debug:")
		qb.PrintSQL = true
		qb.Explain = true
		qb.RunTime = true
		qb.LastQueryCost = true
	}
	qb.disableSQLChecker = true
	raw := qb.SQL(statement)
	qb.disableSQLChecker = false
	core := storager.getCore()
	// PrintSQL
	if qb.PrintSQL {
		debugLog = append(debugLog, renderSQL(qb.debugData.id, raw.Query, raw.Values))
	}
	// EXPLAIN
	if qb.Explain {
		row := core.QueryRowxContext(ctx, "EXPLAIN "+raw.Query, raw.Values...)
		debugLog = append(debugLog, renderExplain(qb.debugData.id, row))
	}
	if qb.RunTime {
		qb.elapsedTimeData.startTime = time.Now()
	}
	if len(debugLog) != 0 {
		Log.Debug(strings.Join(debugLog, "\n"))
	}
	return
}

func (qb *QB) execDebugAfter(ctx context.Context, storager Storager, statement Statement) {
	var err error
	defer func() {
		if err != nil {
			Log.Error("error", "error", err)
		}
	}()
	if qb.RunTime {
		Log.Debug(renderRunTime(qb.debugData.id, time.Now().Sub(qb.elapsedTimeData.startTime)))
	}
	if qb.LastQueryCost {
		var lastQueryCost float64
		lastQueryCost, err = coreLastQueryCost(ctx, storager)
		if err != nil {
			return
		}
		Log.Debug(renderLastQueryCost(qb.debugData.id, lastQueryCost))
	}
}

func (qb QB) whereIsEmpty() bool {
	return qb.Raw.IsZero() && len(qb.Where) == 0 && len(qb.WhereOR) == 0 && qb.WhereRaw.IsZero()
}
