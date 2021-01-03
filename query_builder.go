package sq

import (
	"errors"
	"log"
	"reflect"
	"strings"
)

type Data struct {
	Column Column
	Value interface{}
}

type QB struct {
	Table Tabler
	tableName string
	TableRaw func ()(query string, values []interface{})
	softDelete string
	Select []Column
	Index string
	Where []Condition
	WhereOR [][]Condition
	WhereRaw func ()(query string, values []interface{})
	Update []Data
	Insert []Data
	Limit int
	Join []Join
	Debug bool
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
func (qb QB) SQL(statement Statement) (query string, values []interface{}) {
	var sqlList stringQueue
	if qb.Table == nil && qb.TableRaw == nil {
		panic(errors.New("qb must have qb.Table or qb.TableRaw"))
	}
	if qb.Table != nil {
		qb.tableName = "`" + qb.Table.TableName() + "`"
		qb.softDelete = qb.Table.SoftDelete()
	}
	if qb.TableRaw != nil {
		var subTableValues []interface{}
		qb.tableName, subTableValues = qb.TableRaw()
		values = append(values, subTableValues...)
	}
	statement.Switch(func(_Select int) {
	  sqlList.Push("SELECT")
	  if qb.Table != nil {
	   		rValue := reflect.ValueOf(qb.Table)
	   		rType := rValue.Type()
			if rType.Kind() == reflect.Ptr {
				rValue = rValue.Elem()
				rType = rValue.Type()
			}
	   		for i:=0;i<rType.NumField();i++ {
	   			field := rType.Field(i)
	   			tag, has := field.Tag.Lookup("db")
	   			if !has {continue}
	   			if tag != "" {
					qb.Select = append(qb.Select, Column(tag))
				}
			}
		}
	  if len(qb.Select) == 0 {
	   		sqlList.Push("*")
		} else {
			sqlList.Push(strings.Join(columnsToStringsWithAS(qb.Select), ", "))
		}
		sqlList.Push("FROM")
	  sqlList.Push(qb.tableName)
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
			whereString, whereValues = qb.WhereRaw()
			values = append(values, whereValues...)
		} else {
			tooMuchWhere := len(qb.Where) != 0 && len(qb.WhereOR) != 0
			if tooMuchWhere { panic(errors.New("if qb.WhereOR not empty, then qb.Where must be nil")) }
			if len(qb.Where) != 0 && len(qb.WhereOR) == 0 {
				qb.WhereOR = append(qb.WhereOR, qb.Where)
			}
			var orList stringQueue
			for _, whereAndList := range qb.WhereOR {
				andsQuery, andsValues :=  ToConditions(whereAndList).andsSQL()
				values = append(values, andsValues...)
				orList.Push(andsQuery)
			}
			whereString = orList.Join(") OR (")
			if len(orList.Value) > 1 {
				whereString = "(" + whereString + ")"
			}
		}
		needSoftDelete := qb.softDelete != ""

		if needSoftDelete  {
			whereSofeDelete := qb.softDelete
			if len(whereString) != 0 {
				whereString += " AND " + whereSofeDelete
			} else {
				whereString += whereSofeDelete
			}
		}
		if len(whereString) != 0 {
			sqlList.Push("WHERE")
			sqlList.Push(whereString)
		}
	}

	query = sqlList.Join(" ")
	defer func() {
		if qb.Debug {
			log.Print("goclub/sql debug:\r\n" + query, values)
		}
	}()
	return
}
func (qb QB) SQLSelect() (query string, values []interface{}) {
	return qb.SQL(Statement("").Enum().Select)
}
func (qb QB) SQLInsert() (query string, values []interface{}) {
	return qb.SQL(Statement("").Enum().Insert)
}
func (qb QB) SQLUpdate() (query string, values []interface{}) {
	return qb.SQL(Statement("").Enum().Update)
}