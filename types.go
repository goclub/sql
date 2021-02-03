package sq

import (
	"reflect"
	"strings"
	"time"
)

type ColumnBeforeCreate interface {
	ColumnBeforeCreate()
}
type ColumnBeforeUpdate interface {
	ColumnBeforeUpdate()
}

type CreatedAtUpdatedAt struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
func (v *CreatedAtUpdatedAt) ColumnBeforeCreate() {
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
}
func (v *CreatedAtUpdatedAt) ColumnBeforeUpdate() {
	v.UpdatedAt = time.Now()
}
type CreateTimeUpdateTime struct {
	CreateTime time.Time `db:"create_time"`
	UpdateTime time.Time `db:"update_time"`
}
type GMTCreateGMTUpdate struct {
	GMTCreate time.Time `db:"gmt_create"`
	GMTUpdate time.Time `db:"gmt_update"`
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