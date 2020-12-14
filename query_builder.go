package sq

type UpdateColumn map[Column]interface{}
type OP struct {

}
type Condition struct {
	Column Column
	OP OP
}
type QB struct {
	Table string
	Select []Column
	Where []Condition
	Update UpdateColumn
	Limit int
	Join Join
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
	On []Column
}
type Column string
func Equal(v interface{}) OP {
	return OP{}
}
func GtInt(i int) OP {
	return OP{}
}
type Conditions []Condition

func And(column Column, operator OP) Conditions{
	return Conditions{}.And(column, operator)
}
func (w Conditions) And(column Column, operator OP) Conditions {
	return w
}
func (qb QB) Check(checkSQL ...string) QB {
	return qb
}
func (qb QB) ToSelect() (query string, values []interface{}) {
	return
}

func (qb QB) BindModel(model Model) QB {
	qb.Table = model.TableName()
	return qb
}
func (qb QB) Paging(page int, perPage int) QB {
	return qb
}