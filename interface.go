package sq

import (
	"database/sql"
	"time"
)

type UpdateModeler interface {
	UpdateModelWhere() []Condition
}
type Tabler interface {
	TableName() string
	SoftDeleteWhere() QueryValues
}
type QueryValues struct {
	Query string
	Values []interface{}
}
type table struct {
	tableName string
	softDeleteWhere QueryValues
}
func (t table) TableName() string { return t.tableName }
func (t table) SoftDeleteWhere() (QueryValues) {
	return t.softDeleteWhere
}
func Table(tableName string, softDeleteWhere QueryValues) Tabler {
	return table{
		tableName: tableName,
		softDeleteWhere: softDeleteWhere,
	}
}
type Model interface {
	TableName() string
	SoftDeleteWhere() QueryValues
	SoftDeleteSet() QueryValues
	AfterCreate(result sql.Result) error
	BeforeCreate() error
	BeforeUpdate() error
	AfterUpdate() error
}
type Relation interface {
	TableName() string
	SoftDeleteWhere() QueryValues
	RelationJoin () []Join
}

type SoftDeleteIsDeleted struct {}
func (SoftDeleteIsDeleted) SoftDeleteWhere() (QueryValues) {return QueryValues{"`is_deleted` = 0", nil}}
func (SoftDeleteIsDeleted) SoftDeleteSet() (QueryValues)   {return QueryValues{"`is_deleted` = 1" ,nil}}

type SoftDeleteDeletedAt struct {}
func (SoftDeleteDeletedAt) SoftDeleteWhere() (QueryValues) {return QueryValues{"`deleted_at` IS NULL", nil}}
func (SoftDeleteDeletedAt) SoftDeleteSet() (QueryValues)   {return QueryValues{"`deleted_at` = ?" ,[]interface{}{time.Now()}}}


type DefaultModel struct {
	CreatedAtUpdatedAt
}
func (DefaultModel) AfterCreate(result sql.Result) error {return nil}
func (DefaultModel) BeforeCreate() error {return nil}
func (DefaultModel) BeforeUpdate() error {return nil}
func (DefaultModel) AfterUpdate() error {return nil}
