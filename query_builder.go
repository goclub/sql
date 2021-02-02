package sq

import (
	"errors"
	"log"
	"strings"
)

type Data struct {
	Column Column
	Value interface{}
	Raw Raw
	OnUpdated func() error
}
func Set(column Column, value interface{}) Data {
	return Data{Column: column, Value: value}
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

	Update []Data
	Insert []Data

	Where []Condition
	WhereOR [][]Condition
	WhereRaw func ()Raw

	Limit int
	limitRaw limitRaw
	Offset int

	Lock SelectLock

	Join []Join
	Debug bool
	Raw Raw
	CheckSQL []string
	sqlChecker SQLChecker
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
		column += ` AS "` + s + `"`
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
				sets = append(sets, data.Column.wrapField()+"=?")
				values = append(values, data.Value)
			}
		}
		sqlList.Push(strings.Join(sets, ","))
	}, func(_Delete string) {
		sqlList.Push("DELETE FROM")
		sqlList.Push(qb.tableName)
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
	{
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
	if qb.Offset != 0 {
		sqlList.Push("OFFSET ?")
		values = append(values, qb.Offset)
	}
	if len(qb.Lock) != 0 {
		sqlList.Push(qb.Lock.String())
	}
		query := sqlList.Join(" ")
	defer func() {
		if qb.Debug {
			log.Print("goclub/sql debug:\r\n" + query, "\r\n", values)
		}
		if len(qb.CheckSQL) != 0 {
			matched, diff := qb.sqlChecker.Check(qb.CheckSQL, query)
			if matched == false {
				qb.sqlChecker.Log(diff)
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
	return qb
}