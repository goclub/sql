package sq

import (
	"errors"
	"log"
	"strings"
)

// sq.Set(column, value)
type Update struct {
	Column Column
	Value interface{}
	Raw Raw
	OnUpdated func() error
}
func Set(column Column, value interface{}) Update {
	return Update{Column: column, Value: value}
}
// sq.Value(column, value)
type Insert struct {
	Column Column
	Value interface{}
}
func Value(column Column, value interface{}) Insert {
	return Insert{Column: column, Value: value}
}

type QB struct {
	Table Tabler
		tableName string
	TableRaw TableRaw

	DisableSoftDelete bool
		softDelete Raw

	UnionTable UnionTable

	Select []Column
	SelectRaw []Raw

	Index string

	Update []Update
	// 可使用 sq.Value() 快速创建 sq.Insert []Insert{sq.Value(),sq.Value()}
	Insert []Insert
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
	// string or []string
	CheckSQL interface{}
	SQLChecker SQLChecker
}
func (qb QB) mustInTransaction() error {
	if len(qb.Lock) == 0 {
		return nil
	}
	return errors.New("goclub/sql: SELECT " + qb.Lock.String() + " must exec in transaction")
}

type TableRaw struct {
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
func (s Statement) String() string { return string(s)}
func (Statement) Enum() (e struct {
	Select Statement
	Update Statement
	Delete Statement
	Insert Statement
}){
	e.Select = "SELECT"
	e.Update = "UPDATE"
	e.Delete = "DELETE"
	e.Insert = "INSERT"
	return
}
func (s Statement) Switch(
	Select func(_Select int),
	Update func(_Update bool),
	Delete func(_Delete string),
	Insert func (_Insert []int),
	) {
	enum := s.Enum()
	switch s {
	case enum.Select:
		Select(0)
	case enum.Update:
		Update(false)
	case enum.Delete:
		Delete("")
	case enum.Insert:
		Insert(nil)
	}
}
func (qb QB) SQL(statement Statement) Raw {
	if len(qb.Raw.Query) != 0 {
		return qb.Raw
	}
	var values []interface{}
	var sqlList stringQueue
	if statement == statement.Enum().Select && qb.UnionTable.Tables != nil{
		unionRaw := qb.UnionTable.SQLSelect()
		sqlList.Push(unionRaw.Query)
		values = append(values, unionRaw.Values...)
	}
	if qb.Table != nil {
		qb.tableName = "`" + qb.Table.TableName() + "`"
		switch statement {
		case statement.Enum().Select,
			 statement.Enum().Update:
			qb.softDelete = qb.Table.SoftDeleteWhere()
		case statement.Enum().Insert:
		case statement.Enum().Delete:
		default:
			panic(errors.New("statement can not be " + statement.String()))
		}
	}
	if qb.TableRaw.TableName.Query != ""{
		qb.tableName = qb.TableRaw.TableName.Query
		values = append(values, qb.TableRaw.TableName.Values...)
		qb.softDelete = qb.TableRaw.SoftDeleteWhere
	}
	statement.Switch(func(_Select int) {
	  if qb.UnionTable.Tables == nil {
		  sqlList.Push("SELECT")
		  if qb.SelectRaw == nil {
			  if qb.Table != nil && len(qb.Select) == 0 {
				  qb.Select = TagToColumns(qb.Table)
			  }
			  if len(qb.Select) == 0 {
				  sqlList.Push("*")
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
		  sqlList.Push(qb.tableName)
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
	}, func(_Update bool) {
		sqlList.Push("UPDATE")
		sqlList.Push(qb.tableName)
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
	}, func(_Delete string) {
		sqlList.Push("DELETE FROM")
		sqlList.Push(qb.tableName)
	}, func(_Insert []int) {
			if qb.UseInsertIgnoreInto {
				sqlList.Push("INSERT IGNORE INTO")
			} else {
				sqlList.Push("INSERT INTO")
			}

			sqlList.Push(qb.tableName)
			var columns []string
			for _, item := range qb.Insert {
				columns = append(columns, item.Column.wrapField())
				values = append(values, item.Value)
			}
			sqlList.Push("(" + strings.Join(columns, ",") + ")")
			sqlList.Push("VALUES")
			var placeholders []string
			for _, _ = range columns {
				placeholders = append(placeholders, "?")
			}
			sqlList.Push("(" + strings.Join(placeholders, ",") + ")")
	})
	// where
	{
		var whereString string
		var whereRaw Raw
		if qb.WhereRaw.Query != "" {
			whereRaw = qb.WhereRaw
		} else {
			tooMuchWhere := len(qb.Where) != 0 && len(qb.WhereOR) != 0
			if tooMuchWhere { panic(errors.New("if qb.WhereOR not empty, then qb.Where must be nil")) }
			if len(qb.Where) != 0 && len(qb.WhereOR) == 0 {
				qb.WhereOR = append(qb.WhereOR, qb.Where)
			}
			whereRaw = ConditionsSQL(qb.WhereOR)
		}
		var whereValues []interface{}
		whereString, whereValues = whereRaw.Query, whereRaw.Values
		values = append(values, whereValues...)
		var disableWhereIsEmpty bool
		if statement == statement.Enum().Delete || statement == statement.Enum().Update {
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
	// group by
	if qb.GroupByRaw.Query != "" {
		sqlList.Push("GROUP BY")
		sqlList.Push(qb.GroupByRaw.Query)
		values = append(values, qb.GroupByRaw.Values...)
	} else if len(qb.GroupBy) != 0 {
		sqlList.Push("GROUP BY")
		sqlList.Push(strings.Join(columnsToStrings(qb.GroupBy), ", "))
	}
	if len(qb.Lock) != 0 {
		sqlList.Push(qb.Lock.String())
	}
	// havaing
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
	query := sqlList.Join(" ")
	defer func() {
		if qb.Debug {
			log.Print("goclub/sql debug:\r\n" + query, "\r\n", values)
		}
		if qb.SQLChecker != nil {
			checkSQL := []string{}
			switch v := qb.CheckSQL.(type) {
			case string:
				checkSQL = []string{v}
			case []string:
				checkSQL = v
			}
			matched, diff, stack := qb.SQLChecker.Check(checkSQL, query)
			if matched == false {
				qb.SQLChecker.Log(diff, stack)
			}
		}
	}()
	return Raw{query, values}
}
func (qb QB) SQLSelect() Raw {
	return qb.SQL(Statement("").Enum().Select)
}
func (qb QB) SQLInsert() Raw {
	return qb.SQL(Statement("").Enum().Insert)
}
func (qb QB) SQLUpdate() Raw {
	return qb.SQL(Statement("").Enum().Update)
}
func (qb QB) SQLDelete() Raw {
	return qb.SQL(Statement("").Enum().Delete)
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