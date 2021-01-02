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
