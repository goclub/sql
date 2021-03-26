package sq

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)


type WherePrimaryKeyer interface {
	WherePrimaryKey() []Condition
}
type Tabler interface {
	TableName() string
	SoftDeleteWhere() Raw
	SoftDeleteSet() Raw
}
// 供 relation 使用
type table struct {
	tableName string
	softDeleteWhere func() Raw
	softDeleteSet func() Raw
}
func (t table) TableName() string {
	return t.tableName
}
func (t table) SoftDeleteWhere() Raw {
	return t.softDeleteWhere()
}
func (t table) SoftDeleteSet() Raw {
	return t.softDeleteSet()
}

type Raw struct {
	Query string
	Values []interface{}
}
type Model interface {
	Tabler
	BeforeCreate() error
	AfterCreate(result sql.Result) error
	BeforeUpdate() error
	AfterUpdate() error
}
type Relation interface {
	TableName() string
	SoftDeleteWhere() Raw
	RelationJoin () []Join
}


type WithoutSoftDelete struct {}
func (WithoutSoftDelete) SoftDeleteWhere() Raw {return Raw{}}
func (WithoutSoftDelete) SoftDeleteSet() Raw   {return Raw{}}

type SoftDeleteDeletedAt struct {}
func (SoftDeleteDeletedAt) SoftDeleteWhere() Raw {return Raw{"`deleted_at` IS NULL", nil}}
func (SoftDeleteDeletedAt) SoftDeleteSet() Raw   {return Raw{"`deleted_at` = ?" ,[]interface{}{time.Now()}}}

type SoftDeleteDeleteTime struct {}
func (SoftDeleteDeleteTime) SoftDeleteWhere() Raw {return Raw{"`delete_time` IS NULL", nil}}
func (SoftDeleteDeleteTime) SoftDeleteSet() Raw   {return Raw{"`delete_time` = ?" ,[]interface{}{time.Now()}}}

type SoftDeleteIsDeleted struct {}
func (SoftDeleteIsDeleted) SoftDeleteWhere() Raw {return Raw{"`is_deleted` = 0", nil}}
func (SoftDeleteIsDeleted) SoftDeleteSet() Raw   {return Raw{"`is_deleted` = 1" ,nil}}

type DefaultLifeCycle struct {

}
func (v *DefaultLifeCycle) BeforeCreate() error {return nil}
func (v *DefaultLifeCycle) AfterCreate(result sql.Result) error {return nil}
func (v *DefaultLifeCycle) BeforeUpdate() error {return nil}
func (v *DefaultLifeCycle) AfterUpdate() error {return nil}

type Storager interface {
	getCore() StoragerCore
	getSQLChecker () SQLChecker
}
type StoragerCore interface {
	sqlx.Queryer
	sqlx.QueryerContext
	sqlx.Execer
	sqlx.ExecerContext
	sqlx.Preparer
	sqlx.PreparerContext
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
