package sq

const sqlPlaceholder = "?"
type Condition struct {
	Column Column
	OP OP
}
func ConditionRaw(query string, values []interface{}) Condition {
	return Condition{
		OP: OP{
			RawQuery: query,
			Values: values,
		},
	}
}
func And(column Column, operator OP) conditions{
	return conditions{}.And(column, operator)
}
func AndRaw(query string, values []interface{}) (column Column, operator OP) {
	return "", OP{
		RawQuery: query,
		Values: values,
	}
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
func (w conditions) andsSQL() Raw {
	var andList stringQueue
	var values []interface{}
	for _, c :=  range w {
		andList.Push(c.OP.sql(c.Column, &values))
	}
	query := andList.Join(" AND ")
	return Raw{query, values}
}