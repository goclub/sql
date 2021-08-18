package m

import (
	sq "github.com/goclub/sql"
)

type UserWithAddress struct {
	UserID            IDUser  `db:"user.id"`
	Name          string  `db:"user.name"`
	Mobile        string  `db:"user.mobile"`
	ChinaIDCardNo string  `db:"user.china_id_card_no"`
	Address       string  `db:"user_address.address"`
}

func (a UserWithAddress) SoftDeleteWhere() sq.Raw {
	return sq.Raw{"`user`.`deleted_at` IS NULL", nil}
}

func (a UserWithAddress) RelationJoin() []sq.Join {
	return []sq.Join{
		{
			Type: 	  	   sq.LeftJoin,
			TableName:	   "user_address",
			On:"`user`.`id` = `user_address`.`user_id`",
		},
	}
}

func (UserWithAddress) TableName() string {
	return "user"
}

func (v UserWithAddress) Column() (col struct{
	UserID         sq.Column
	Name           sq.Column
	Mobile         sq.Column
	ChinaIDCardNo  sq.Column
	Address        sq.Column
}) {
	col.UserID             = "user.id"
	col.Name           = "user.name"
	col.Mobile         = "user.mobile"
	col.ChinaIDCardNo  = "user.china_id_card_no"
	col.Address = "user_address.address"
	return
}


func buildCheck() {
	var v sq.Relation
	v = &UserWithAddress{}
	_=v
}