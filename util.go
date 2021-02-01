package sq

import (
	"strings"
)

type stringQueue struct {
	Value []string
}
func (v *stringQueue) Push(args... string) {
	v.Value = append(v.Value, args...)
}
func (v stringQueue) Join(sep string) string {
	return strings.Join(v.Value, sep)
}
type stringQueueBindValue struct {
	Value string
	Has bool
}
func (sList *stringQueue) PopBind(last *stringQueueBindValue) stringQueue {
	listLen := len(sList.Value)
	if listLen == 0 {
		/*
			Clear StringListBindValue Because in this case
				```
				list.PopBind(&last)
				// do Something..
				list.PopBind(&last)
				```
				last test same var
		*/
		last.Value = ""
		last.Has = false
		return *sList
	}
	last.Value = sList.Value[listLen-1]
	last.Has = true
	sList.Value = sList.Value[:listLen-1]
	return *sList
}
func columnsToStrings (columns []Column) (strings []string) {
	for _, column := range columns {
		strings = append(strings, column.wrapField())
	}
	return
}
func columnsToStringsWithAS (columns []Column) (strings []string) {
	for _, column := range columns {
		strings = append(strings, column.wrapFieldWithAS())
	}
	return
}
