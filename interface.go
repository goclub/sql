package sq

import (
	"database/sql"
)
type Table interface {
	TableName() string
	SoftDelete() string
}
type Model interface {
	TableName() string
	SoftDelete() string
	BeforeCreate()
	AfterCreate(result sql.Result) error
}

type Relation interface {
	FormTable () string
	RelationJoin () []Join
}