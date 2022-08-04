package sq

import (
	"context"
	"github.com/jmoiron/sqlx"
)

// sq.Table("user",nil, nil)
// sq.Table("user", sq.Raw{"`deleted_at` IS NULL", nil}, sq.Raw{"`deleted_at` = ?" ,[]interface{}{time.Now()}})
func Table(tableName string, softDeleteWhere func() Raw, softDeleteSet func() Raw) Tabler {
	if softDeleteWhere == nil {
		softDeleteWhere = func() Raw {
			return Raw{}
		}
	}
	if softDeleteSet == nil {
		softDeleteSet = func() Raw {
			return Raw{}
		}
	}
	return table{
		tableName:       tableName,
		softDeleteWhere: softDeleteWhere,
		softDeleteSet:   softDeleteSet,
	}
}

type Tabler interface {
	TableName() string
	SoftDeleteWhere() Raw
	SoftDeleteSet() Raw
}

// 供 relation sq.Table() 使用
type table struct {
	tableName       string
	softDeleteWhere func() Raw
	softDeleteSet   func() Raw
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
	Query  string
	Values []interface{}
}
type Model interface {
	Tabler
	BeforeInsert() error
	AfterInsert(result Result) error
	BeforeUpdate() error
	AfterUpdate() error
}
type Relation interface {
	TableName() string
	SoftDeleteWhere() Raw
	RelationJoin() []Join
}

type DefaultLifeCycle struct {
}

func (v *DefaultLifeCycle) BeforeInsert() error                 { return nil }
func (v *DefaultLifeCycle) AfterInsert(result Result) error { return nil }
func (v *DefaultLifeCycle) BeforeUpdate() error                 { return nil }
func (v *DefaultLifeCycle) AfterUpdate() error                  { return nil }

type Storager interface {
	getCore() StoragerCore
	getSQLChecker() SQLChecker
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

type sqlInsertRawer interface {
	SQLInsertRaw() (raw Raw)
}