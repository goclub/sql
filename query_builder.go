package sq


type OP struct {}
type Condition struct {
	Column Column
	OP OP
}
type QB struct {
	Table string
	Select []Column
	Where []Condition
	Check []string
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
