package sq
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
	}
}
func GtInt(i int) OP {
	return OP {
		Symbol: ">",
	}
}