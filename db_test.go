package sq_test

import (
	"context"
	sq "github.com/goclub/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

var testDB *sq.Database
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
	sq.ExecMigrate(db, &Migrate{})
}

func TestDB(t *testing.T) {
	suite.Run(t, new(TestDBSuite))
}
type TestDBSuite struct {
	suite.Suite
}
func insertUser(id IDUser, name string, age int) {
	_, err := testDB.Core.Exec("INSERT INTO `user` (`id`, `name`, `age`) VALUES (?,?,?)", id, name, age)
	if err != nil {
		panic(err)
	}
}
func (suite TestDBSuite) TestQueryRowScan() {
	t := suite.T()
	id := sq.UUID()
	insertUser(IDUser(id), "DB_QueryRowScan", 1)
	userCol := User{}.Column()
	{
		var name string
		var age int
		has ,err := testDB.QueryRowScan(context.TODO(), sq.QB{
			Table: User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.ID , sq.Equal(id)),
		}, &name, &age)
		assert.Equal(t, has, true)
		assert.Equal(t, err , nil)
		assert.Equal(t, name, "DB_QueryRowScan")
		assert.Equal(t, age, 1)
	}
	{
		var name string
		var age int
		err := testDB.Transaction(context.TODO(), func(tx *sq.Transaction) sq.TxResult {
			has ,err := tx.QueryRowScan(context.TODO(), sq.QB{
				Table: User{},
				Select: []sq.Column{userCol.Name, userCol.Age},
				Where: sq.And(userCol.ID , sq.Equal(id)),
			}, &name, &age)
			assert.Equal(t, has, true)
			assert.Equal(t, err , nil)
			assert.Equal(t, name, "DB_QueryRowScan")
			assert.Equal(t, age, 1)
			return tx.Commit()
		})
		assert.Equal(t, err , nil)
	}
}
func (suite TestDBSuite) TestCreateModel() {
	user := User{
		Name:"nimo",
	}
	err := testDB.CreateModel(context.TODO(), &user) ; if err != nil {
		panic(err)
	}
}
func (suite TestDBSuite) TestRelation() {
	// userWithAddress := UserWithAddress{}
	// userWithAddressCol := userWithAddress.Column()
	// has, err := testDB.QueryRelation(context.TODO(), &userWithAddress, sq.QB{
	// 	Debug:true,
	// 	Where: sq.And(userWithAddressCol.Name, sq.Equal("nimo")),
	// }) ; if err != nil {
	// 	panic(err)
	// }
	// log.Print(has, userWithAddress)
}

