package sq

// 递增存在并发问题，现在不满意目前设计的接口，等以后在想怎么设计 2021年02月02日10:32:15 @nimoc。
// // 递减可能存在并发问题，所以只在事务中处理
// func (tx *Transaction) DecrementIntModel(ctx context.Context, ptr Model, props IncrementInt, checkSQL ...string) (affected bool, err error) {
// 	field := props.Column.wrapField()
// 	result, err := coreModelUpdate(ctx, tx.Core, ptr, []Data{
// 		{
// 			// SET age = age - ?
// 			Raw: Raw{
// 				Query: field + " = " + field + " - ?",
// 				Values: []interface{}{props.Value},
// 			},
// 			OnUpdated: func() error {
// 				return props.OnUpdated(props.Value)
// 			},
// 		},
// 	}, []Condition{
// 		ConditionRaw(
// 			// WHERE age  >= stock
// 			field + " >= " + props.AfterIncrementLessThanOrEqual.wrapField(),
// 			[]interface{}{props.Value},
// 		),
// 	})
// 	if err != nil {
// 		return
// 	}
// 	rowsAffected, err := result.RowsAffected() ; if err != nil {
// 		return
// 	}
// 	affected = rowsAffected !=0
// 	return
// }
// type DecrementFloat struct {
// 	Column Column
// 	Value float64
// 	AfterDecrementGreaterThanOrEqual Column
// 	OnUpdated func(value float64) error
// }
// // 递减可能存在并发问题，所以只在事务中处理
// func (tx *Transaction) DecrementFloatModel(ctx context.Context, ptr Model, props IncrementFloat, checkSQL ...string) (affected bool, err error) {
// 	field := props.Column.wrapField()
// 	result, err := coreModelUpdate(ctx, tx.Core, ptr, []Data{
// 		{
// 			// SET age = age - ?
// 			Raw: Raw{
// 				Query: field + " = " + field + " - ?",
// 				Values: []interface{}{props.Value},
// 			},
// 			OnUpdated: func() error {
// 				return props.OnUpdated(props.Value)
// 			},
// 		},
// 	}, []Condition{
// 		ConditionRaw(
// 			// WHERE age  >= stock
// 			field + " >= " + props.AfterIncrementLessThanOrEqual.wrapField(),
// 			[]interface{}{props.Value},
// 		),
// 	})
// 	if err != nil {
// 		return
// 	}
// 	rowsAffected, err := result.RowsAffected() ; if err != nil {
// 		return
// 	}
// 	affected = rowsAffected !=0
// 	return
// }
// type DecrementInt struct {
// 	Column Column
// 	Value uint
// 	AfterDecrementGreaterThanOrEqual Column
// 	OnUpdated func(value uint) error
// }
// type IncrementInt struct {
// 	Column Column
// 	Value uint
// 	AfterIncrementLessThanOrEqual Column
// 	OnUpdated func(value uint) error
// }
// // 递增可能存在并发问题，所以只在事务中处理
// func (tx *Transaction) IncrementIntModel(ctx context.Context, ptr Model, props IncrementInt) (affected bool, err error) {
// 	field := props.Column.wrapField()
// 	result, err := db.ModelUpdate(ctx, ptr, []Data{
// 		{
// 			// SET age = age + ?
// 			Raw: Raw{
// 				Query: field + " = " + field + " + ?",
// 				Values: []interface{}{props.Value},
// 			},
// 			OnUpdated: func() error {
// 				return props.OnUpdated(props.Value)
// 			},
// 		},
// 	}, []Condition{
// 		ConditionRaw(
// 			// WHERE age + ? <= stock
// 			field + " + ? <= " + props.AfterIncrementLessThanOrEqual.wrapField(),
// 			[]interface{}{props.Value},
// 		),
// 	})
// 	if err != nil {
// 		return
// 	}
// 	rowsAffected, err := result.RowsAffected() ; if err != nil {
// 		return
// 	}
// 	affected = rowsAffected !=0
// 	return
// }
// type IncrementFloat struct {
// 	Column Column
// 	Value float64
// 	AfterIncrementLessThanOrEqual Column
// 	OnUpdated func(value float64) error
// }
// func (tx *Transaction) IncrementFloatModel(ctx context.Context, ptr Model, props IncrementFloat) (affected bool, err error) {
// 	field := props.Column.wrapField()
// 	result, err := tx.ModelUpdate(ctx, ptr, []Data{
// 		{
// 			// SET age = age + ?
// 			Raw: Raw{
// 				Query: field + " = " + field + " + ?",
// 				Values: []interface{}{props.Value},
// 			},
// 			OnUpdated: func() error {
// 				return props.OnUpdated(props.Value)
// 			},
// 		},
// 	}, []Condition{
// 		ConditionRaw(
// 			// WHERE age + ? <= stock
// 			field + " + ? <= " + props.AfterIncrementLessThanOrEqual.wrapField(),
// 			[]interface{}{props.Value},
// 		),
// 	})
// 	if err != nil {
// 		return
// 	}
// 	rowsAffected, err := result.RowsAffected() ; if err != nil {
// 		return
// 	}
// 	affected = rowsAffected !=0
// 	return
// }
