package sq

import (
	"errors"
	"reflect"
	"time"
)

type OP struct {
	Symbol string
	Values []interface{}
	Raw string
	Placeholder string
}
func (op OP) sql(column Column) string {
	var and stringQueue
	if len(op.Raw) != 0 {
		and.Push(op.Raw)
	} else {
		and.Push(column.wrapField())
		and.Push(op.Symbol)
		if len(op.Placeholder) != 0 {
			and.Push(op.Placeholder)
		} else {
			and.Push(sqlPlaceholder)
		}
	}
	return and.Join(" ")
}
func Equal(v interface{}) OP {
	return OP {
		Symbol: "=",
		Values: []interface{}{v},
	}
}
func NotEqual(v interface{}) OP {
	return OP{
		Symbol: "<>",
		Values: []interface{}{v},
	}
}
func Raw(raw string, values ...interface{}) OP {
	return OP {
		Raw: raw,
		Values: values,
	}
}
func SubQuery(symbol string, qb QB) OP {
	query, values := qb.SQLSelect()
	return OP {
		Placeholder: "(" + query + ")",
		Symbol: symbol,
		Values: values,
	}
}
func Like(s string) OP {
	return OP {
		Symbol: "LIKE",
		Values: []interface{}{"%" + s + "%"},
	}
}
func In(slice interface{}) OP {
	var placeholder string
	var values []interface{}
	rValue := reflect.ValueOf(slice)
	if rValue.Type().Kind() != reflect.Slice {
		panic(errors.New("sq.In(" + rValue.Type().Name() + ") slice must be slice"))
	}
	if rValue.Len() == 0 {
		placeholder = "(NULL)"
	} else {
		var placeholderList stringQueue
		for i:=0;i<rValue.Len();i++ {
			values = append(values, rValue.Index(i).Interface())
			placeholderList.Push(sqlPlaceholder)
		}
		placeholder = "(" + placeholderList.Join(", ") + ")"
	}
	return OP{
		Symbol:      "IN",
		Values:      values,
		Placeholder: placeholder,
	}
}
func LikeLeft(s string) OP {
	return OP {
		Symbol: "LIKE",
		Values: []interface{}{s + "%"},
	}
}
func LikeRight(s string) OP {
	return OP {
		Symbol: "LIKE",
		Values: []interface{}{"%" + s},
	}
}
func GtInt(i int) OP {
	return OP {
		Symbol: ">",
		Values: []interface{}{i},
	}
}
func GtOrEqualInt(i int) OP {
	return OP {
		Symbol: ">=",
		Values: []interface{}{i},
	}
}
func LtInt(i int) OP {
	return OP {
		Symbol: "<",
		Values: []interface{}{i},
	}
}
func LtOrEqualInt(i int) OP {
	return OP {
		Symbol: "<=",
		Values: []interface{}{i},
	}
}
func GtFloat(i float64) OP {
	return OP {
		Symbol: ">",
		Values: []interface{}{i},
	}
}
func GtOrEqualFloat(i float64) OP {
	return OP {
		Symbol: ">=",
		Values: []interface{}{i},
	}
}
func LtFloat(i float64) OP {
	return OP {
		Symbol: "<",
		Values: []interface{}{i},
	}
}
func LtOrEqualFloat(i float64) OP {
	return OP {
		Symbol: "<=",
		Values: []interface{}{i},
	}
}
func GtTime(t time.Time) OP {
	return OP {
		Symbol: ">",
		Values: []interface{}{t},
	}
}
func GtOrEqualTime(t time.Time) OP {
	return OP {
		Symbol: ">=",
		Values: []interface{}{t},
	}
}
func LtTime(t time.Time) OP {
	return OP {
		Symbol: "<",
		Values: []interface{}{t},
	}
}
func LtOrEqualTime(t time.Time) OP {
	return OP {
		Symbol: "<=",
		Values: []interface{}{t},
	}
}