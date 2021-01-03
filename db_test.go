package sq_test

import (
	"context"
	sq "github.com/goclub/sql"
	"github.com/stretchr/testify/suite"
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
	}.String()) ; if err != nil {
		panic(err)
	}
	testDB = db
	_=dbClose // init 场景下不需要 close，应该在 main 执行完毕后 close
	err = testDB.Core.Ping() ; if err != nil {
		panic(err)
	}
}
func TestDB(t *testing.T) {
	suite.Run(t, new(TestDBSuite))
}
type TestDBSuite struct {
	suite.Suite
}

func (suite TestDBSuite) TestCreateModel() {
	user := User{
		Name:"nimo",
	}
	err := testDB.CreateModel(context.TODO(), &user) ; if err != nil {
		panic(err)
	}
	log.Print(user)
}
func (suite TestDBSuite) TestRelation() {
	userWithAddress := UserWithAddress{}
	userWithAddressCol := userWithAddress.Column()
	has, err := testDB.Relation(context.TODO(), &userWithAddress, sq.QB{
		Where: sq.And(userWithAddressCol.Name, sq.Equal("nimo")),
		Debug: true,
	}) ; if err != nil {
		panic(err)
	}
	log.Print(has, userWithAddress)
}

func (suite TestDBSuite) TestUpdate() {

}
