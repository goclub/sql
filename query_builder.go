package sq

import (
	xerr "github.com/goclub/error"
	"log"
	"runtime/debug"
	"strings"
)

// sq.Set(column, value)
type Update struct {
	Column Column
	Value interface{}
	Raw Raw
}
type updates []Update
func (u updates) Set(column Column, value interface{}) updates {
	if op, ok := value.(OP); ok {
		value = op.Values[0]
		DefaultLog.Print("sq.Set(column, value) value can not be sq.Equal(v) or sq.OP{}, may be you need use like sq.Set(\"id\", taskID)")
		DefaultLog.Print(string(debug.Stack()))
	}
	u = append(u, Update{
		Column: column,
		Value: value,
	})
	return u
}
func (u updates) SetRaw(query string, values ...interface{}) updates {
	for i, value := range values {
		if op, ok := value.(OP); ok {
			values[i] = op.Values[0]
			DefaultLog.Print("goclub/sql: sq.SetRaw(query, values) values element can not be sq.Equal(v) or sq.OP{}, may be you need use like sq.Set(\"id\", taskID)")
			DefaultLog.Print(string(debug.Stack()))
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
	Value interface{}
}

type QB struct {
	Select []Column
	SelectRaw []Raw

	From Tabler
		from string
	FromRaw FromRaw

	DisableSoftDelete bool
		softDelete Raw

	UnionTable UnionTable

	Index string

	Update []Update
	// UPDATE IGNORE
	UseUpdateIgnore bool
	Insert Values
	InsertMultiple InsertMultiple
	// INSERT IGNORE INTO
	UseInsertIgnoreInto bool

	Where []Condition
	WhereOR [][]Condition
	WhereRaw Raw

	OrderBy []OrderBy
	OrderByRaw Raw

	GroupBy []Column
	GroupByRaw Raw

	Having []Condition
	HavingRaw Raw

	Limit int
	limitRaw limitRaw
	Offset int

	Lock SelectLock

	Join []Join
	Raw Raw

	Debug bool
	Review string
	Reviews []string
	SQLChecker SQLChecker

}
func (qb QB) mustInTransaction() error {
	if len(qb.Lock) == 0 {
		return nil
	}
	return xerr.New("goclub/sql: SELECT " + qb.Lock.String() + " must exec in transaction")
}

type FromRaw struct {
	TableName Raw
	SoftDeleteWhere Raw
}
type OrderBy struct {
	Column Column
	Type orderByType
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
	Tables []QB
	UnionAll bool
}
func (union UnionTable) SQLSelect() (raw Raw) {
	var sqlList stringQueue
	var subQueryList []string
	for _, table := range union.Tables {
		subQV := table.SQLSelect()
		subQueryList = append(subQueryList, "(" + subQV.Query + ")")
		raw.Values = append(raw.Values, subQV.Values...)
	}
	unionText := "UNION"
	if union.UnionAll {
		unionText += " ALL"
	}
	sqlList.Push(strings.Join(subQueryList, " "+ unionText+ " "))
	raw.Query = sqlList.Join(" ")
	return
}
type limitRaw struct {
	Valid bool
	Limit int
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
	Type JoinType
	TableName string
	On string
}
type Column string
func (c Column) String() string { return string(c)}
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

func (s Statement) String() string { return string(s)}

func (qb QB) SQL(statement Statement) Raw {
	if len(qb.Raw.Query) != 0 {
		return qb.Raw
	}
	var values []interface{}
	var sqlList stringQueue
	if statement == StatementSelect && qb.UnionTable.Tables != nil{
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
	if qb.FromRaw.TableName.Query != ""{
		qb.from = qb.FromRaw.TableName.Query
		values = append(values, qb.FromRaw.TableName.Values...)
		qb.softDelete = qb.FromRaw.SoftDeleteWhere
	}
	switch statement {
	case StatementSelect:
		if qb.UnionTable.Tables == nil {
			sqlList.Push("SELECT")
			if qb.SelectRaw == nil {
				if qb.From != nil && len(qb.Select) == 0 {
					qb.Select = TagToColumns(qb.From)
				}
				if len(qb.Select) == 0 {
					return Raw{Query: "goclub/sql: if qb.SelectRaw is nil or qb.Form is nil then qb.Select can not be nil or empty slice"}
				} else {
					sqlList.Push(strings.Join(columnsToStringsWithAS(qb.Select), ", "))
				}
			} else{
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
		var sets  []string
		for _, data := range qb.Update {
			if len(data.Raw.Query) !=0  {
				sets = append(sets, data.Raw.Query)
				values = append(values, data.Raw.Values...)
			} else {
				sets = append(sets, data.Column.wrapField()+"= ?")
				values = append(values, data.Value)
			}
		}
		sqlList.Push(strings.Join(sets, ","))
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
				rowPlaceholder = append(rowPlaceholder, "?")
				values = append(values, v)
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
			if tooMuchWhere { panic(xerr.New("if qb.WhereOR not empty, then qb.Where must be nil")) }
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
			if needSoftDelete  {
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
	if qb.HavingRaw.Query != ""{
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
		if qb.Debug {
			DefaultLog.Printf("goclub/sql: debug\n%s\n%#+v", query, values)
		}
		if qb.Review != "" {
			qb.Reviews = append(qb.Reviews, qb.Review)
		}
		if len(qb.Reviews) != 0 {
			matched, refs, err := qb.SQLChecker.Check(qb.Reviews, query) ; if err != nil {
				qb.SQLChecker.TrackFail(err, qb.Reviews, query, "")
			} else {
				if matched == false {
					qb.SQLChecker.TrackFail(err, qb.Reviews, query, refs)
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
func (qb QB) Paging(page int, perPage int) QB {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 10
		log.Print("goclub/sql: Paging(page, perPage) alert perPage is 0 ,perPage can't be 0 . gofree will set perPage 10. but you need check your code.")
	}
	qb.Offset = (page - 1) * perPage
	qb.Limit = perPage
	return qb
}