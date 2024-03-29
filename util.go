package sq

import (
	"database/sql"
	xerr "github.com/goclub/error"
	"reflect"
	"strings"
)

type stringQueue struct {
	Value []string
}

func (v *stringQueue) Push(args ...string) {
	v.Value = append(v.Value, args...)
}
func (v stringQueue) Join(sep string) string {
	return strings.Join(v.Value, sep)
}

type stringQueueBindValue struct {
	Value string
	Has   bool
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
func columnsToStrings(columns []Column) (strings []string) {
	for _, column := range columns {
		strings = append(strings, column.wrapField())
	}
	return
}
func columnsToStringsWithAS(columns []Column) (strings []string) {
	for _, column := range columns {
		strings = append(strings, column.wrapFieldWithAS())
	}
	return
}

type primaryIDInfo struct {
	HasID   bool
	IDValue interface{}
}

func CheckRowScanErr(scanErr error) (has bool, err error) {
	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return false, nil
		} else {
			return false, scanErr
		}
	} else {
		has = true
	}
	return
}

func PlaceholderSlice(slice interface{}) (placeholder string) {
	var values []interface{}
	rValue := reflect.ValueOf(slice)
	if rValue.Type().Kind() != reflect.Slice {
		panic(xerr.New("sq.PlaceholderIn(" + rValue.Type().Name() + ") slice must be slice"))
	}
	if rValue.Len() == 0 {
		placeholder = "(NULL)"
	} else {
		var placeholderList []string
		for i := 0; i < rValue.Len(); i++ {
			values = append(values, rValue.Index(i).Interface())
			placeholderList = append(placeholderList, "?")
		}
		placeholder = "(" + strings.Join(placeholderList, ",") + ")"
	}
	return
}
