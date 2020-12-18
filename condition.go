package sq

const sqlPlaceholder = "?"
type Condition struct {
	Column Column
	OP OP
}
func And(column Column, operator OP) conditions{
	return conditions{}.And(column, operator)
}
func ToConditions(c []Condition) conditions {
	return conditions(c)
}
type conditions []Condition
func (w conditions) And(column Column, operator OP) conditions {
	w = append(w, Condition{
		Column: column,
		OP: operator,
	})
	return w
}
func (w conditions) andsSQL() (query string, values []interface{}) {
	var andList stringQueue
	for _, c :=  range w {
		andList.Push(c.OP.sql(c.Column))
		values = append(values, c.OP.Values...)
	}
	query = andList.Join(" AND ")
	return
}