package sq

import (
	xerr "github.com/goclub/error"
	"reflect"
	"strings"
)

type Tag struct {
	Value string
}
func (t Tag) IsIgnoreInsert() bool {
	sqTags := strings.Split(t.Value, "|")
	for _, tag := range sqTags {
		if strings.TrimSpace(tag) == "ignoreInsert" {
			return true
		}
	}
	return false
}
func (t Tag) IsIgnoreUpdate() bool {
	sqTags := strings.Split(t.Value, "|")
	for _, tag := range sqTags {
		if strings.TrimSpace(tag) == "ignoreUpdate" {
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
		panic(xerr.New("goclub/sql: Too many structures are nested"))
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