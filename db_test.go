package sq_test

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	sq "github.com/goclub/sql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
	"time"
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

func (suite TestDBSuite) TestInsert() {
	t := suite.T()
	userCol := User{}.Column()
	newID := sq.UUID()
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.Like("TestInsert")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		result, err := testDB.Insert(context.TODO(), sq.QB{
			Table: TableUser{},
			Insert: []sq.Insert{
				sq.Value(userCol.ID, newID),
				sq.Value(userCol.Name, "TestInsert"),
				sq.Value(userCol.Age, 18),
			},
			CheckSQL:[]string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.QueryModel(context.TODO(), &user, sq.QB{
			CheckSQL: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestInsert")
		assert.Equal(t, user.Age, 18)
		assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
}

func (suite TestDBSuite) TestCreateModel() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCreateModel")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	var userID IDUser
	{
		user := User{
			Name: "TestCreateModel",
			Age: 18,
		}
		err := testDB.CreateModel(
			context.TODO(),
			&user,
			"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)",
		)
		userID = user.ID
		assert.NoError(t, err)
		assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
	{
		user := User{}
		has, err := testDB.QueryModel(context.TODO(), &user, sq.QB{
			CheckSQL: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(userID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, userID)
		assert.Equal(t, user.Name, "TestCreateModel")
		assert.Equal(t, user.Age, 18)
		assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
}




func (suite TestDBSuite) TestQueryRowScan() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestQueryRowScan")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	// 插入数据
	{
		user := User{Name:"TestQueryRowScan", Age: 20,}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		var name string
		var age uint64
		has, err := testDB.QueryRowScan(context.TODO(), sq.QB{
			Table: User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.Equal("TestQueryRowScan")),
			CheckSQL: []string{"SELECT `name`, `age` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL LIMIT ?"},
		}, &name, &age)
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, name, "TestQueryRowScan")
		assert.Equal(t, age, uint64(20))
	}
	{
		var name string
		var age uint64
		has, err := testDB.QueryRowScan(context.TODO(), sq.QB{
			Table: User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.Equal("TestQueryRowScanNotExist")),
			CheckSQL: []string{"SELECT `name`, `age` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL LIMIT ?"},
		}, &name, &age)
		assert.NoError(t, err)
		assert.Equal(t, has, false)
		assert.Equal(t, name, "")
		assert.Equal(t, age, uint64(0))
	}
}



func (suite TestDBSuite) QueryRowStructScan() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("QueryRowStructScan")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	// 插入数据
	{
		user := User{Name:"QueryRowStructScan", Age: 20,}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		type Data struct {
			Name string `db:"name"`
			Age int `db:"age"`
		}
		var data Data
		has, err := testDB.QueryRowStructScan(context.TODO(), &data, sq.QB{
			Table: User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestQueryRowScan")),
			CheckSQL: []string{"SELECT `name`, `age` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, data, Data{"QueryRowStructScan", 20})
	}
	{
		type Data struct {
			Name string `db:"name"`
			Age int `db:"age"`
		}
		var data Data
		has, err := testDB.QueryRowStructScan(context.TODO(), &data, sq.QB{
			Table: User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.Equal("TestQueryRowScanNotExist")),
			CheckSQL: []string{"SELECT `name`, `age` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
		assert.Equal(t, data, Data{})
	}
}
func (suite TestDBSuite) TestSelectScan() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSelectScan")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	// 插入数据
	{
		user := User{Name:"TestSelectScan_1", Age: 20,}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		user := User{Name:"TestSelectScan_2", Age: 21,}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		type Data struct {
			Name string `db:"name"`
			Age int `db:"age"`
		}
		var list []Data
		err := testDB.SelectScan(context.TODO(), sq.QB{
			Table: User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSelectScan")),
			OrderBy: []sq.OrderBy{{userCol.Name, sq.ASC}},
			CheckSQL: []string{"SELECT `name`, `age` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL ORDER BY `name` ASC"},
		}, func(rows *sqlx.Rows) error {
			data := Data{}
			err := rows.StructScan(&data) ; if err != nil {
				return err
			}
			list = append(list, data)
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, list, []Data{
			{"TestSelectScan_1", 20},
			{"TestSelectScan_2", 21},
		})
	}
}

func (suite TestDBSuite) TestSelectSlice() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSelectSlice")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	// 插入数据
	{
		user := User{Name:"TestSelectSlice_1", Age: 20,}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		user := User{Name:"TestSelectSlice_2", Age: 21,}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		type Data struct {
			Name string `db:"name"`
			Age int `db:"age"`
		}
		var list []Data
		err := testDB.SelectSlice(context.TODO(), &list, sq.QB{
			Table: User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSelectSlice")),
			OrderBy: []sq.OrderBy{{userCol.Name, sq.ASC}},
			CheckSQL: []string{"SELECT `name`, `age` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL ORDER BY `name` ASC"},
		},)
		assert.NoError(t, err)
		assert.Equal(t, list, []Data{
			{"TestSelectSlice_1", 20},
			{"TestSelectSlice_2", 21},
		})
	}
}
func (suite TestDBSuite) TestCount() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCount")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCount")),
		})
		assert.NoError(t, err)
		assert.Equal(t, count, 0)
	}
	// 插入数据
	{
		user := User{Name:"TestCount_1"}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCount")),
		})
		assert.NoError(t, err)
		assert.Equal(t, count, 1)
	}
	{
		user := User{Name:"TestCount_2"}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCount")),
		})
		assert.NoError(t, err)
		assert.Equal(t, count, 2)
	}
}

func (suite TestDBSuite) TestHas() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestHas")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestHas")),
			CheckSQL: []string{"SELECT 1 FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
	}
	// 插入数据
	{
		user := User{Name:"TestHas_1"}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestHas")),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
	}
}



func (suite TestDBSuite) TestSum() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		value, err := testDB.Sum(context.TODO(), userCol.Age, sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			CheckSQL: []string{"SELECT SUM(`age`) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, value, sql.NullInt64{
			Int64: 0,
			Valid: false,
		})
	}
	// 插入数据
	{
		user := User{Name:"TestSum_1"}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		value, err := testDB.Sum(context.TODO(), userCol.Age, sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			CheckSQL: []string{"SELECT SUM(`age`) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, value, sql.NullInt64{
			Int64: 0,
			Valid: true,
		})
	}
	// 插入数据
	{
		user := User{Name:"TestSum_2", Age: 20}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		value, err := testDB.Sum(context.TODO(), userCol.Age, sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			CheckSQL: []string{"SELECT SUM(`age`) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, value, sql.NullInt64{
			Int64: 20,
			Valid: true,
		})
	}
	// 插入数据
	{
		user := User{Name:"TestSum_3", Age: 20}
		err := testDB.CreateModel(context.TODO(), &user)
		assert.NoError(t, err)
	}
	{
		value, err := testDB.Sum(context.TODO(), userCol.Age, sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			CheckSQL: []string{"SELECT SUM(`age`) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, value, sql.NullInt64{
			Int64: 40,
			Valid: true,
		})
	}
}


func (suite TestDBSuite) TestQueryModel() {
	t := suite.T()
	userCol := User{}.Column()
	newID := sq.UUID()
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.Like("TestQueryModel")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		result, err := testDB.Insert(context.TODO(), sq.QB{
			Table: TableUser{},
			Insert: []sq.Insert{
				sq.Value(userCol.ID, newID),
				sq.Value(userCol.Name, "TestQueryModel"),
				sq.Value(userCol.Age, 18),
			},
			CheckSQL:[]string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.QueryModel(context.TODO(), &user, sq.QB{
			CheckSQL: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestQueryModel")
		assert.Equal(t, user.Age, 18)
		assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
}


func (suite TestDBSuite) TestQueryModelSlice() {
	t := suite.T()
	userCol := User{}.Column()

	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table:    User{},
			Where:    sq.And(userCol.Name, sq.Like("TestQueryModelSlice")),
			CheckSQL: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		var users  []User
		err := testDB.QueryModelSlice(context.TODO(), &users, sq.QB{
			Where: sq.And(userCol.Name, sq.Like("TestQueryModelSlice")),
		})
		assert.NoError(t, err)
		assert.Equal(t, len(users), 0)
	}
	{
		for i:=0;i<10;i++ {
			result, err := testDB.Insert(context.TODO(), sq.QB{
				Table: TableUser{},
				Insert: []sq.Insert{
					sq.Value(userCol.ID, sq.UUID()),
					sq.Value(userCol.Name, "TestQueryModelSlice_" + strconv.Itoa(i)),
					sq.Value(userCol.Age, i),
				},
				CheckSQL:[]string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
			})
			assert.NoError(t, err)
			affected, err := result.RowsAffected()
			assert.NoError(t, err)
			assert.Equal(t, affected, int64(1))
		}
	}
	{
		 users := []User{}
		 err := testDB.QueryModelSlice(context.TODO(), &users, sq.QB{
			CheckSQL: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL ORDER BY `name` ASC"},
			Where: sq.And(userCol.Name, sq.Like("TestQueryModelSlice")),
			OrderBy: []sq.OrderBy{{userCol.Name, sq.ASC}},
		})
		assert.NoError(t, err)
		 for i:=0;i<10;i++ {
		 	user := users[i]
			 assert.NoError(t, err)
			 assert.Equal(t, len(user.ID), 36)
			 assert.Equal(t, user.Name, "TestQueryModelSlice_" + strconv.Itoa(i))
			 assert.Equal(t, user.Age, i)
			 assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
			 assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
		 }
	}
}


func (suite TestDBSuite) TestUpdate() {
	t := suite.T()
	userCol := User{}.Column()
	newID := sq.UUID()
	createTime := time.Now()
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.Like("TestUpdate")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		result, err := testDB.Insert(context.TODO(), sq.QB{
			Table: TableUser{},
			Insert: []sq.Insert{
				sq.Value(userCol.ID, newID),
				sq.Value(userCol.Name, "TestUpdate"),
				sq.Value(userCol.Age, 18),
			},
			CheckSQL:[]string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.QueryModel(context.TODO(), &user, sq.QB{
			CheckSQL: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestUpdate")
		assert.Equal(t, user.Age, 18)
		assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
	time.Sleep(time.Second)
	{
		result, err := testDB.Update(context.TODO(), sq.QB{
				Table: User{},
				Where: sq.And(userCol.ID, sq.Equal(newID)),
				Update: []sq.Update{
					sq.Set(userCol.Name, "TestUpdate_changed"),
				},
				CheckSQL: []string{"UPDATE `user` SET `name`=? WHERE `id` = ? AND `deleted_at` IS NULL"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.QueryModel(context.TODO(), &user, sq.QB{
			CheckSQL: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestUpdate_changed")
		assert.Equal(t, user.Age, 18)
		assert.True(t, createTime.Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
}


func (suite TestDBSuite) TestUpdateModel() {
	t := suite.T()
	userCol := User{}.Column()
	newID := sq.UUID()
	createTime := time.Now()
	{
		_, err := testDB.HardDelete(context.TODO(), sq.QB{
			Table: User{},
			Where: sq.And(userCol.Name, sq.Like("TestUpdateModel")),
			CheckSQL:[]string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		result, err := testDB.Insert(context.TODO(), sq.QB{
			Table: TableUser{},
			Insert: []sq.Insert{
				sq.Value(userCol.ID, newID),
				sq.Value(userCol.Name, "TestUpdateModel"),
				sq.Value(userCol.Age, 18),
			},
			CheckSQL:[]string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.QueryModel(context.TODO(), &user, sq.QB{
			CheckSQL: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestUpdateModel")
		assert.Equal(t, user.Age, 18)
		assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
	time.Sleep(time.Second)
	{
		user := User{
			ID: IDUser(newID),
			Name: "",
		}
		result, err := testDB.UpdateModel(context.TODO(), &user, []sq.Update{
			sq.Set(userCol.Name, "TestUpdateModel_changed"),
		}, nil, "UPDATE `user` SET `name`=? WHERE `id` = ? AND `deleted_at` IS NULL",
		)
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
		assert.Equal(t, user.Name, "TestUpdateModel_changed")
	}
	{
		user := User{}
		has, err := testDB.QueryModel(context.TODO(), &user, sq.QB{
			CheckSQL: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestUpdateModel_changed")
		assert.Equal(t, user.Age, 18)
		assert.True(t, createTime.Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
}




