package sq

import (
	"errors"
	"log"
	"strings"
)

type Data struct {
	Column Column
	Value interface{}
}

type QB struct {
	Union Union
	Table Tabler
		tableName string
	TableRaw QueryValues
	DisableSoftDelete bool
		softDelete QueryValues
	Select []Column
	SelectRaw []QueryValues
	Index string
	Where []Condition
	WhereOR [][]Condition
	WhereRaw func ()QueryValues
	Update []Data
	Insert []Data
	Limit int
	limitRaw limitRaw
	Join []Join
	Debug bool
}
type Union struct {
	Tables []QB
	UnionAll bool
}
func (union Union) SQLSelect() (qv QueryValues) {
	var sqlList stringQueue
	var subQueryList []string
	for _, table := range union.Tables {
		subQV := table.SQLSelect()
		subQueryList = append(subQueryList, "(" + subQV.Query + ")")
		qv.Values = append(qv.Values, subQV.Values...)
	}
	unionText := "UNION"
	if union.UnionAll {
		unionText += " ALL"
	}
	sqlList.Push(strings.Join(subQueryList, " "+ unionText+ " "))
	qv.Query = sqlList.Join(" ")
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
		column += ` AS "` + s + `"`
	}
	return column
}
func (qb QB) Check(checkSQL ...string) QB {
	return qb
}

func (qb QB) Paging(page int, perPage int) QB {
	return qb
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
func (qb QB) SQL(statement Statement) QueryValues {
	var values []interface{}
	var sqlList stringQueue
	if statement == statement.Enum().Select && qb.Union.Tables != nil{
		unionQueryValues := qb.Union.SQLSelect()
		sqlList.Push(unionQueryValues.Query)
		values = append(values, unionQueryValues.Values...)
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
	if qb.TableRaw.Query != ""{
		qb.tableName = qb.TableRaw.Query
		values = append(values, qb.TableRaw.Values...)
	}
	statement.Switch(func(_Select int) {
	  if qb.Union.Tables == nil {
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
			  for _, queryValues := range qb.SelectRaw {
				  rawColumns = append(rawColumns, queryValues.Query)
				  values = append(values, queryValues.Values...)
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
			sets = append(sets, data.Column.wrapField()+"=?")
			values = append(values, data.Value)
		}
		sqlList.Push(strings.Join(sets, ","))
	}, func(_Delete string) {

	}, func(_Insert []int) {
			sqlList.Push("INSERT INTO")
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
	if statement == statement.Enum().Select || statement == statement.Enum().Update {
		var whereString string
		if qb.WhereRaw != nil {
			var whereValues []interface{}
			whereQV := qb.WhereRaw()
			whereString, whereValues = whereQV.Query, whereQV.Values
			values = append(values, whereValues...)
		} else {
			tooMuchWhere := len(qb.Where) != 0 && len(qb.WhereOR) != 0
			if tooMuchWhere { panic(errors.New("if qb.WhereOR not empty, then qb.Where must be nil")) }
			if len(qb.Where) != 0 && len(qb.WhereOR) == 0 {
				qb.WhereOR = append(qb.WhereOR, qb.Where)
			}
			var orList stringQueue
			for _, whereAndList := range qb.WhereOR {
				andsQV := ToConditions(whereAndList).andsSQL()
				values = append(values, andsQV.Values...)
				orList.Push(andsQV.Query)
			}
			whereString = orList.Join(") OR (")
			if len(orList.Value) > 1 {
				whereString = "(" + whereString + ")"
			}
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
	limit := qb.Limit
	// 优先使用 qb.limitRaw, 因为 db.Count 需要用到
	if qb.limitRaw.Valid {
		limit = qb.limitRaw.Limit
	}
	if limit != 0 {
		sqlList.Push("LIMIT ?")
		values = append(values, qb.Limit)
	}
	query := sqlList.Join(" ")
	defer func() {
		if qb.Debug {
			log.Print("goclub/sql debug:\r\n" + query, "\r\n", values)
		}
	}()
	return QueryValues{query, values}
}
func (qb QB) SQLSelect() QueryValues {
	return qb.SQL(Statement("").Enum().Select)
}
func (qb QB) SQLInsert() QueryValues {
	return qb.SQL(Statement("").Enum().Insert)
}
func (qb QB) SQLUpdate() QueryValues {
	return qb.SQL(Statement("").Enum().Update)
}