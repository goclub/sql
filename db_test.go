package sq_test

import (
	"context"
	sq "github.com/goclub/sql"
	"log"
	"testing"
)

var testDB *sq.DB
func init () {
	db, dbClose, err := sq.Open("mysql", sq.DataSourceName{
		DriverName: "mysql",
		User: "root",
		Password:"somepass",
		Host: "127.0.0.1",
		Port:"3306",
		DB: "test_goclub_sql",
		Query: map[string]string{
			"charset": "utf8",
			"parseTime": "true",
			"loc": "Local",
		},
	}.String()) ; if err != nil {
		panic(err)
	}
	testDB = db
	_=dbClose // init 场景下不需要 close，应该在 main 执行完毕后 close
	err = testDB.Core.Ping() ; if err != nil {
		panic(err)
	}
}
func TestDateTime(t *testing.T) {
	var user User
	userCol := user.Column()
	has, err := testDB.QueryRowStructScan(context.TODO(), &user, sq.QB{
		Table:    User{},
		Where: sq.And(userCol.ID, sq.Equal("0d2d88ab-4035-11eb-a5e1-0242ac120003")),
	})
	log.Print(user.CreatedAt.String(),user.UpdatedAt.String())
	log.Print(has, err)
	log.Printf("%+v", user)
}

