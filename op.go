package sq

import (
	xerr "github.com/goclub/error"
	"reflect"
)

type OP struct {
	Query string
	Values []interface{}
	Symbol string
	Placeholder string
	Multiple []OP
	OrGroup []Condition
	Ignore bool
}
func (op OP) sql(column Column, values *[]interface{}) string {
	var and stringQueue
	if len(op.OrGroup) != 0 {
		raw := conditions(op.OrGroup).coreSQL("OR")
		if raw.Query == "" {
			return ""
		}
		raw.Query = "(" + raw.Query + ")"
		*values = append(*values, raw.Values...)
		return raw.Query
	} else if len(op.Multiple) != 0 {
		for _, subOP := range op.Multiple {
			and.Push(subOP.sql(column, values))
		}
	} else {
		if len(op.Query) != 0 {
			and.Push(op.Query)
			*values = append(*values, op.Values...)
		} else {
			and.Push(column.wrapField())
			and.Push(op.Symbol)
			if len(op.Placeholder) != 0 {
				and.Push(op.Placeholder)
			} else {
				and.Push(sqlPlaceholder)
			}
			*values = append(*values, op.Values...)
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
func SubQuery(symbol string, qb QB) OP {
	raw := qb.SQLSelect()
	query, values := raw.Query, raw.Values
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
		panic(xerr.New("sq.In(" + rValue.Type().Name() + ") slice must be slice"))
	}
	if rValue.Len() == 0 {
		placeholder = "(NULL)"
	} else {
		var placeholderList stringQueue
		for i:=0;i<rValue.Len();i++ {
			values = append(values, rValue.Index(i).Interface())
			placeholderList.Push(sqlPlaceholder)
		}
		placeholder = "(" + placeholderList.Join(",") + ")"
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
func Between(begin interface{}, end interface{}) OP {
	return OP {
		Symbol: "BETWEEN",
		Values: []interface{}{begin, end},
		Placeholder: `? AND ?`,
	}
}
func NotBetween(begin interface{}, end interface{}) OP {
	return OP {
		Symbol: "NOT BETWEEN",
		Values: []interface{}{begin, end},
		Placeholder: `? AND ?`,
	}
}
func GT(v interface{}) OP {
	return OP {
		Symbol: ">",
		Values: []interface{}{v},
	}
}
func GTE(v interface{}) OP {
	return OP {
		Symbol: ">=",
		Values: []interface{}{v},
	}
}
func LT(v interface{}) OP {
	return OP {
		Symbol: "<",
		Values: []interface{}{v},
	}
}
func LTE(v interface{}) OP {
	return OP {
		Symbol: "<=",
		Values: []interface{}{v},
	}
}

func IsNull() OP {
	return OP{
		Symbol: "IS NULL",
		Values: nil,
	}
}
func Multiple(ops []OP) OP {
	return OP{
		Multiple: ops,
	}
}
func IF(condition bool, op OP) OP {
	if condition == false {
		op.Ignore = true
	}
	return op
}