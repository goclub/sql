package sq

import "time"

type WithoutSoftDelete struct{}

func (WithoutSoftDelete) SoftDeleteWhere() Raw { return Raw{} }
func (WithoutSoftDelete) SoftDeleteSet() Raw   { return Raw{} }

type SoftDeletedAt struct{}

func (SoftDeletedAt) SoftDeleteWhere() Raw { return Raw{"`deleted_at` IS NULL", nil} }
func (SoftDeletedAt) SoftDeleteSet() Raw   { return Raw{"`deleted_at` = ?", []interface{}{time.Now()}} }

type SoftDeleteTime struct{}

func (SoftDeleteTime) SoftDeleteWhere() Raw { return Raw{"`delete_time` IS NULL", nil} }
func (SoftDeleteTime) SoftDeleteSet() Raw   { return Raw{"`delete_time` = ?", []interface{}{time.Now()}} }

type SoftIsDeleted struct{}

func (SoftIsDeleted) SoftDeleteWhere() Raw { return Raw{"`is_deleted` = 0", nil} }
func (SoftIsDeleted) SoftDeleteSet() Raw   { return Raw{"`is_deleted` = 1", nil} }
