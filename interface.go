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
	BeforeCreate() error
	AfterCreate(result sql.Result) error
	BeforeUpdate() error
	AfterUpdate() error
}
type Relation interface {
	TableName() string
	SoftDeleteWhere() QueryValues
	RelationJoin () []Join
}

type SoftDeleteDeletedAt struct {}
func (SoftDeleteDeletedAt) SoftDeleteWhere() (QueryValues) {return QueryValues{"`deleted_at` IS NULL", nil}}
func (SoftDeleteDeletedAt) SoftDeleteSet() (QueryValues)   {return QueryValues{"`deleted_at` = ?" ,[]interface{}{time.Now()}}}

type SoftDeleteDeleteTime struct {}
func (SoftDeleteDeleteTime) SoftDeleteWhere() (QueryValues) {return QueryValues{"`delete_time` IS NULL", nil}}
func (SoftDeleteDeleteTime) SoftDeleteSet() (QueryValues)   {return QueryValues{"`delete_time` = ?" ,[]interface{}{time.Now()}}}

type SoftDeleteIsDeleted struct {}
func (SoftDeleteIsDeleted) SoftDeleteWhere() (QueryValues) {return QueryValues{"`is_deleted` = 0", nil}}
func (SoftDeleteIsDeleted) SoftDeleteSet() (QueryValues)   {return QueryValues{"`is_deleted` = 1" ,nil}}

type DefaultLifeCycle struct {

}
func (v *DefaultLifeCycle) BeforeCreate() error {return nil}
func (v *DefaultLifeCycle) AfterCreate(result sql.Result) error {return nil}
func (v *DefaultLifeCycle) BeforeUpdate() error {return nil}
func (v *DefaultLifeCycle) AfterUpdate() error {return nil}
