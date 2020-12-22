package sq

import (
	"time"
)

type OP struct {
	Symbol string
	Values []interface{}
	Raw string
	SubQuery string
}
func (op OP) sql(column Column) string {
	var and stringQueue
	if len(op.Raw) != 0 {
		and.Push(op.Raw)
	} else {
		and.Push(column.wrapField())
		and.Push(op.Symbol)
		if len(op.SubQuery) != 0 {
			and.Push(op.SubQuery)
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
		SubQuery: "(" + query + ")",
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