package sq

import "database/sql"

type Tabler interface {
	TableName() string
	SoftDelete() string
}
type table struct {
	tableName string
	softDelete string
}
func (t table) TableName() string { return t.tableName }
func (t table) SoftDelete() string { return t.softDelete }

func Table(tableName string, softDelete string) table {
	return table{
		tableName: tableName,
		softDelete:softDelete,
	}
}
type Model interface {
	TableName() string
	SoftDelete() string
	AfterCreate(result sql.Result) error
	BeforeCreate()
	BeforeUpdate()
}
type Relation interface {
	TableName () string
	SoftDelete() string
	RelationJoin () []Join
}