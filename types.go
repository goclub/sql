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