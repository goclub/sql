package sq

const sqlPlaceholder = "?"
type Condition struct {
	Column Column
	OP OP
}
func ConditionRaw(query string, values []interface{}) Condition {
	return Condition{
		OP: OP{
			Query: query,
			Values: values,
		},
	}
}
func And(column Column, operator OP) conditions{
	return conditions{}.And(column, operator)
}
func AndRaw(query string, values []interface{}) conditions {
	return And("", OP{
		Query: query,
		Values: values,
	})
}
func OrGroup(conditions ...Condition) conditions {
	op := OP{OrGroup: conditions,}
	item :=  Condition{OP:op}
	return []Condition{item}
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
func (w conditions) AndRaw(query string, values []interface{}) conditions {
	w = append(w, Condition{
		Column: "",
		OP: OP{
			Query: query,
			Values: values,
		},
	})
	return w
}
func (w conditions) OrGroup(conditions []Condition) conditions {
	op := OP{OrGroup: conditions}
	item := Condition{OP:op}
	w = append(w, item)
	return w
}
func (w conditions) sql(split string) Raw {
	var andList stringQueue
	var values []interface{}
	for _, c :=  range w {
		if c.OP.Ignore {
			continue
		}
		sql := c.OP.sql(c.Column, &values)
		if len(sql) != 0 {
			andList.Push(sql)
		}
	}
	query := andList.Join(" "+ split + " ")
	return Raw{query, values}
}