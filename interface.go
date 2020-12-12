package sq

import "context"

type Model interface {
	TableName() string
	BeforeCreate()
}
type design interface {
	Create(ctx context.Context, ptr Model) error
}