package sq

import (
	"errors"
	"strings"
)

type UpdateColumn map[Column]interface{}


type QB struct {
	Table string
	TableRaw func ()(query string, values []interface{})
	Select []Column
	Where []Condition
	WhereOR [][]Condition
	WhereRaw func ()(query string, values []interface{})
	Update UpdateColumn
	Limit int
	Join Join
	SoftDelete Column
}
const DisableSoftDelete Column = "DisableSoftDelete"
const DefaultSoftDeletedField = "deleted_at"
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
	On []Column
}
type Column string
func (c Column) String() string { return string(c)}
func (c Column) wrapField() string {
	return "`" + strings.ReplaceAll(c.String(), ".", "`.`") + "`"
}
func (qb QB) Check(checkSQL ...string) QB {
	return qb
}

func (qb QB) BindModel(model Model) QB {
	qb.Table = model.TableName()
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
	tableName := "`" + qb.Table + "`"
	if qb.TableRaw != nil {
		var subTableValues []interface{}
		tableName, subTableValues = qb.TableRaw()
		values = append(values, subTableValues...)
	}
	statement.Switch(
	   func(_Select int) {
	   	sqlList.Push("SELECT")
	   	if len(qb.Select) == 0 {
	   		sqlList.Push("*")
		} else {
			sqlList.Push("`" + strings.Join(columnsToStrings(qb.Select), "`, `") + "`")
		}
		sqlList.Push("FROM")
	   	sqlList.Push(tableName)
	}, func(_Update bool) {

	}, func(_Delete string) {

	}, func(_Insert []int) {

	})
	// where
	{
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
		needSoftDelete := true
		switch qb.SoftDelete {
		case "":
			qb.SoftDelete = DefaultSoftDeletedField
		case DisableSoftDelete:
			qb.SoftDelete = ""
			needSoftDelete = false
		}
		if needSoftDelete  {
			whereSofeDelete := qb.SoftDelete.wrapField() + " IS NULL"
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
	return
}
func (qb QB) SQLSelect() (query string, values []interface{}) {
	return qb.SQL(Statement("").Enum().Select)
}