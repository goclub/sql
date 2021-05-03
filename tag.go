package sq

import (
	"errors"
	"reflect"
	"strings"
)

type Tag struct {
	Value string
}
func (t Tag) IsIgnore() bool {
	sqTags := strings.Split(t.Value, "|")
	for _, tag := range sqTags {
		if tag == "ignore" {
			return true
		}
	}
	return false
}
func TagToColumns(v interface{}) (columns []Column) {
	rValue := reflect.ValueOf(v)
	rType := rValue.Type()
	if rType.Kind() == reflect.Ptr {
		rValue = rValue.Elem()
		rType = rValue.Type()
	}
	tier := 0
	scanTagToColumns(rValue, rType, &columns, &tier)
	return
}
func scanTagToColumns(rValue reflect.Value, rType reflect.Type, columns *[]Column, tier *int) {
	if *tier > 10 {
		panic(errors.New("goclub/sql: Too many structures are nested"))
	}
	for i:=0;i<rType.NumField();i++ {
		structField := rType.Field(i)
		tag, has := structField.Tag.Lookup("db")
		if !has {
			if structField.Type.Kind() == reflect.Struct {
				fieldValue := rValue.Field(i)
				scanTagToColumns(fieldValue, structField.Type, columns, tier)
				// fieldValue
			}
			continue
		}
		if tag != "" {
			*columns = append(*columns, Column(tag))
		}
	}
}