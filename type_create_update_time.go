package sq

import (
	"reflect"
	"strings"
	"time"
)

type CreatedAtUpdatedAt struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
type CreateTimeUpdateTime struct {
	CreateTime time.Time `db:"create_time"`
	UpdateTime time.Time `db:"update_time"`
}
type GMTCreateGMTModified struct {
	GMTCreate time.Time `db:"gmt_create"`
	GMTModified time.Time `db:"gmt_modified"`
}
func setTimeNow (fieldValue reflect.Value, fieldType reflect.StructField) {
	if fieldValue.IsZero() {
		if fieldType.Type.String() == "time.Time" {
			now := time.Now()
			if strings.HasPrefix(fieldType.Name, "GMT") {
				now = now.In(time.UTC)
			}
			fieldValue.Set(reflect.ValueOf(now))
		}
	}
}