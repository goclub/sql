package sq_test

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	xerr "github.com/goclub/error"
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
	// sq.DefaultLog = log.New(bytes.NewBuffer(nil), "", log.Lshortfile)
	db, dbClose, err := sq.Open("mysql", sq.MysqlDataSource{
		User: "root",
		Password:"somepass",
		Host: "127.0.0.1",
		Port:"3306",
		DB: "test_goclub_sql",
	}.FormatDSN()) ; if err != nil {
		panic(err)
	}
	testDB = db
	_=dbClose // init 场景下不需要 close，应该在 main 执行完毕后 close
	err = testDB.Core.Ping() ; if err != nil {
		panic(err)
	}
	sq.ExecMigrate(db, &Migrate{})
	_, err = testDB.Exec(context.TODO(), "TRUNCATE TABLE user", nil) ; if err != nil {
		panic(err)
	}
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
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &TableUser{},
			Where: sq.And(userCol.Name, sq.Like("TestInsert")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: sq.Values{
				{userCol.ID, newID},
				{userCol.Name, "TestInsert"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.Query(context.TODO(), &user, sq.QB{
			Reviews: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestInsert")
		assert.Equal(t, user.Age, 18)
		// assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		// assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
}

func (suite TestDBSuite) TestInsertModel() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestInsertModel")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	var userID IDUser
	{
		user := User{
			Name: "TestInsertModel",
			Age: 18,
		}
		_, err := testDB.InsertModel(
			context.TODO(),
			&user,
			sq.QB{
				// Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`,`created_at`,`updated_at`) VALUES (?,?,?,?,?)"}
			},
		)
		userID = user.ID
		assert.NoError(t, err)
		assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
	{
		user := User{}
		has, err := testDB.Query(context.TODO(), &user, sq.QB{
			Reviews: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(userID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, userID)
		assert.Equal(t, user.Name, "TestInsertModel")
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
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestQueryRowScan")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	// 插入数据
	{
		user := User{Name:"TestQueryRowScan", Age: 20,}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		var name string
		var age uint64
		has, err := testDB.QueryRowScan(context.TODO(), sq.QB{
			From: &User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.Equal("TestQueryRowScan")),
			Reviews: []string{"SELECT `name`, `age` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL LIMIT ?"},
		}, []interface{}{&name, &age})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, name, "TestQueryRowScan")
		assert.Equal(t, age, uint64(20))
	}
	{
		var name string
		var age uint64
		has, err := testDB.QueryRowScan(context.TODO(), sq.QB{
			From: &User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.Equal("TestQueryRowScanNotExist")),
			Reviews: []string{"SELECT `name`, `age` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL LIMIT ?"},
		}, []interface{}{&name, &age})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
		assert.Equal(t, name, "")
		assert.Equal(t, age, uint64(0))
	}
}



func (suite TestDBSuite) TestQuery() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestQueryRowScan")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	// 插入数据
	{
		user := User{Name:"TestQueryRowScan", Age: 20,}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		type Data struct {
			Name string `db:"name"`
			Age int `db:"age"`
			TableUser
		}
		{
			var data Data
			has, err := testDB.Query(context.TODO(), &data, sq.QB{
				Where: sq.And(userCol.Name, sq.LikeLeft("TestQueryRowScan")),
				Reviews: []string{"SELECT `name`, `age` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
			})
			assert.NoError(t, err)
			assert.Equal(t, has, true)
			assert.Equal(t, data, Data{Name: "TestQueryRowScan", Age: 20})
		}
		// 测试自定义select覆盖自动 select
		{
			var data Data
			has, err := testDB.Query(context.TODO(), &data, sq.QB{
				Select: []sq.Column{"name"},
				Where: sq.And(userCol.Name, sq.LikeLeft("TestQueryRowScan")),
				Reviews: []string{"SELECT `name` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
			})
			assert.NoError(t, err)
			assert.Equal(t, has, true)
			assert.Equal(t, data, Data{Name: "TestQueryRowScan", Age: 0})
		}
	}
	{
		type Data struct {
			Name string `db:"name"`
			Age int `db:"age"`
			TableUser
		}
		var data Data
		has, err := testDB.Query(context.TODO(), &data, sq.QB{
			Where: sq.And(userCol.Name, sq.Equal("TestQueryRowScanNotExist")),
			Reviews: []string{"SELECT `name`, `age` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
		assert.Equal(t, data, Data{})
	}
}
func (suite TestDBSuite) TestQuerySliceScaner() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestQuerySliceScaner")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	// 插入数据
	{
		user := User{Name:"TestQuerySliceScaner_1", Age: 20,}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		user := User{Name:"TestQuerySliceScaner_2", Age: 21,}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		type Data struct {
			Name string `db:"name"`
			Age int `db:"age"`
		}
		var list []Data
		err := testDB.QuerySliceScaner(context.TODO(), sq.QB{
			From: &User{},
			Select: []sq.Column{userCol.Name, userCol.Age},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestQuerySliceScaner")),
			OrderBy: []sq.OrderBy{{userCol.Name, sq.ASC}},
			Reviews: []string{"SELECT `name`, `age` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL ORDER BY `name` ASC"},
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
			{"TestQuerySliceScaner_1", 20},
			{"TestQuerySliceScaner_2", 21},
		})
	}
}

func (suite TestDBSuite) TestQuerySlice() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestQuerySlice")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	// 插入数据
	{
		user := User{Name:"TestQuerySlice_1", Age: 20,}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		user := User{Name:"TestQuerySlice_2", Age: 21,}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		type Data struct {
			Name string `db:"name"`
			Age int `db:"age"`
			TableUser
		}
		var list []Data
		err := testDB.QuerySlice(context.TODO(), &list, sq.QB{
			Where: sq.And(userCol.Name, sq.LikeLeft("TestQuerySlice"),),
			OrderBy: []sq.OrderBy{{userCol.Name, sq.ASC}},
			Reviews: []string{"SELECT `name`, `age` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL ORDER BY `name` ASC"},
		},)
		assert.NoError(t, err)
		assert.Equal(t, list, []Data{
			{Name: "TestQuerySlice_1", Age: 20,},
			{Name: "TestQuerySlice_2", Age: 21,},
		})
	}
}
func (suite TestDBSuite) TestCount() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCount")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCount")),
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(0))
	}
	// 插入数据
	{
		user := User{Name:"TestCount_1"}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCount")),
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(1))
	}
	{
		user := User{Name:"TestCount_2"}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestCount")),
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(2))
	}
}

func (suite TestDBSuite) TestHas() {
	t := suite.T()
	userCol := User{}.Column()
	// 清空数据
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestHas")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestHas")),
			Reviews: []string{"SELECT 1 FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
	}
	// 插入数据
	{
		user := User{Name:"TestHas_1"}
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestHas")),
			Review: " SELECT 1 FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL LIMIT ?",
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
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		value, err := testDB.Sum(context.TODO(), userCol.Age, sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			Reviews: []string{"SELECT SUM(`age`) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
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
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		value, err := testDB.Sum(context.TODO(), userCol.Age, sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			Reviews: []string{"SELECT SUM(`age`) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
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
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		value, err := testDB.Sum(context.TODO(), userCol.Age, sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			Reviews: []string{"SELECT SUM(`age`) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
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
		_, err := testDB.InsertModel(context.TODO(), &user, sq.QB{})
		assert.NoError(t, err)
	}
	{
		value, err := testDB.Sum(context.TODO(), userCol.Age, sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSum")),
			Reviews: []string{"SELECT SUM(`age`) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
		})
		assert.NoError(t, err)
		assert.Equal(t, value, sql.NullInt64{
			Int64: 40,
			Valid: true,
		})
	}
}


// func (suite TestDBSuite) TestQueryModel() {
// 	t := suite.T()
// 	userCol := User{}.Column()
// 	newID := sq.UUID()
// 	{
// 		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
// 			From: &User{},
// 			Where: sq.And(userCol.Name, sq.Like("TestQueryModel")),
// 			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
// 		})
// 		assert.NoError(t, err)
// 		result, err := testDB.Insert(context.TODO(), sq.QB{
// 			From: &TableUser{},
// 			Insert: []sq.Insert{
// 				{userCol.ID, newID},
// 				{userCol.Name, "TestQueryModel"},
// 				{userCol.Age, 18},
// 			},
// 			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
// 		})
// 		assert.NoError(t, err)
// 		affected, err := result.RowsAffected()
// 		assert.NoError(t, err)
// 		assert.Equal(t, affected, int64(1))
// 	}
// 	{
// 		user := User{}
// 		has, err := testDB.Query(context.TODO(), &user, sq.QB{
// 			Reviews: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
// 			Where: sq.And(userCol.ID, sq.Equal(newID)),
// 		})
// 		assert.NoError(t, err)
// 		assert.Equal(t, has, true)
// 		assert.Equal(t, user.ID, IDUser(newID))
// 		assert.Equal(t, user.Name, "TestQueryModel")
// 		assert.Equal(t, user.Age, 18)
// 		assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
// 		assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
// 	}
// }


func (suite TestDBSuite) TestQueryModelSlice() {
	t := suite.T()
	userCol := User{}.Column()

	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From:    &User{},
			Where:    sq.And(userCol.Name, sq.Like("TestQueryModelSlice")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		var users  []User
		err := testDB.QuerySlice(context.TODO(), &users, sq.QB{
			Where: sq.And(userCol.Name, sq.Like("TestQueryModelSlice")),
		})
		assert.NoError(t, err)
		assert.Equal(t, len(users), 0)
	}
	{
		for i:=0;i<10;i++ {
			result, err := testDB.Insert(context.TODO(), sq.QB{
				From: &TableUser{},
				Insert: sq.Values{
					{userCol.ID, sq.UUID()},
					{userCol.Name, "TestQueryModelSlice_" + strconv.Itoa(i)},
					{userCol.Age, i},
				},
				Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
			})
			assert.NoError(t, err)
			affected, err := result.RowsAffected()
			assert.NoError(t, err)
			assert.Equal(t, affected, int64(1))
		}
	}
	{
		 users := []User{}
		 err := testDB.QuerySlice(context.TODO(), &users, sq.QB{
			Reviews: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL ORDER BY `name` ASC"},
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
			 // assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
			 // assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
		 }
	}
}


func (suite TestDBSuite) TestUpdate() {
	t := suite.T()
	userCol := User{}.Column()
	newID := sq.UUID()
	createTime := time.Now()
	_=createTime
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestUpdate")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: []sq.Insert{
				{userCol.ID, newID},
				{userCol.Name, "TestUpdate"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.Query(context.TODO(), &user, sq.QB{
			Reviews: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestUpdate")
		assert.Equal(t, user.Age, 18)
		// assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		// assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
	time.Sleep(time.Second)
	{
		result, err := testDB.Update(context.TODO(), sq.QB{
				From: &User{},
				Where: sq.And(userCol.ID, sq.Equal(newID)),
				Update: sq.Set(userCol.Name, "TestUpdate_changed"),
				Reviews: []string{"UPDATE `user` SET `name`= ? WHERE `id` = ? AND `deleted_at` IS NULL"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.Query(context.TODO(), &user, sq.QB{
			Reviews: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestUpdate_changed")
		assert.Equal(t, user.Age, 18)
		// assert.True(t, createTime.Sub(user.CreatedAt) < time.Second)
		// assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
}


func (suite TestDBSuite) TestUpdateModel() {
	t := suite.T()
	userCol := User{}.Column()
	newID := sq.UUID()
	createTime := time.Now()
	_=createTime

	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestUpdateModel")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: sq.Values{
				{userCol.ID, newID},
				{userCol.Name, "TestUpdateModel"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		user := User{}
		has, err := testDB.Query(context.TODO(), &user, sq.QB{
			Reviews: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
			Where: sq.And(userCol.ID, sq.Equal(newID)),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, user.ID, IDUser(newID))
		assert.Equal(t, user.Name, "TestUpdateModel")
		assert.Equal(t, user.Age, 18)
		// assert.True(t, time.Now().Sub(user.CreatedAt) < time.Second)
		// assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	}
	time.Sleep(time.Second)
	// {
	// 	user := User{
	// 		ID: IDUser(newID),
	// 		Name: "",
	// 	}
	// 	result, err := testDB.UpdateModel(context.TODO(), &user, []sq.Update{
	// 		sq.Set(userCol.Name, "TestUpdateModel_changed"),
	// 	}, sq.QB{
	// 		Review: "UPDATE `user` SET `name`= ? WHERE `id` = ? AND `deleted_at` IS NULL",
	// 	},
	// 	)
	// 	assert.NoError(t, err)
	// 	affected, err := result.RowsAffected()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, affected, int64(1))
	// 	assert.Equal(t, user.Name, "TestUpdateModel_changed")
	// }
	// {
	// 	user := User{}
	// 	has, err := testDB.Query(context.TODO(), &user, sq.QB{
	// 		Reviews: []string{"SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"},
	// 		Where: sq.And(userCol.ID, sq.Equal(newID)),
	// 	})
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, has, true)
	// 	assert.Equal(t, user.ID, IDUser(newID))
	// 	assert.Equal(t, user.Name, "TestUpdateModel_changed")
	// 	assert.Equal(t, user.Age, 18)
	// 	// assert.True(t, createTime.Sub(user.CreatedAt) < time.Second)
	// 	// assert.True(t, time.Now().Sub(user.UpdatedAt) < time.Second)
	// }
}

func (suite TestDBSuite) TestHardDelete() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestHardDelete")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)

	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From:             &User{},
			DisableSoftDelete: true,
			Where:             sq.And(userCol.Name, sq.LikeLeft("TestHardDelete")),
			Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(0))
	}
	{
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: []sq.Insert{
				{userCol.ID, sq.UUID()},
				{userCol.Name, "TestHardDelete"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From:             &User{},
			DisableSoftDelete: true,
			Where:             sq.And(userCol.Name, sq.LikeLeft("TestHardDelete")),
			Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(1))
	}
	{
		result, err := testDB.HardDelete(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestHardDelete")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From:             &User{},
			DisableSoftDelete: true,
			Where:             sq.And(userCol.Name, sq.LikeLeft("TestHardDelete")),
			Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(0))
	}
}





func (suite TestDBSuite) TestHardDeleteModel() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestHardDeleteModel")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)

	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From:             &User{},
			DisableSoftDelete: true,
			Where:             sq.And(userCol.Name, sq.LikeLeft("TestHardDeleteModel")),
			Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(0))
	}
	newID := sq.UUID()
	{
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: []sq.Insert{
				{userCol.ID, newID},
				{userCol.Name, "TestHardDeleteModel"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From:             &User{},
			DisableSoftDelete: true,
			Where:             sq.And(userCol.Name, sq.LikeLeft("TestHardDeleteModel")),
			Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(1))
	}
	// {
	// 	result, err := testDB.HardDeleteModel(context.TODO(), &User{ID:IDUser(newID)}, sq.QB{Reviews: []string{"DELETE FROM `user` WHERE `id` = ? LIMIT ?"}})
	// 	assert.NoError(t, err)
	// 	affected, err := result.RowsAffected()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, affected, int64(1))
	// }
	// {
	// 	count, err := testDB.Count(context.TODO(), sq.QB{
	// 		From:             &User{},
	// 		DisableSoftDelete: true,
	// 		Where:             sq.And(userCol.Name, sq.LikeLeft("TestHardDeleteModel")),
	// 		Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ?"},
	// 	})
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, count, uint64(0))
	// }
}



func (suite TestDBSuite) TestSoftDelete() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestSoftDelete")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)

	}
	{
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: []sq.Insert{
				{userCol.ID, sq.UUID()},
				{userCol.Name, "TestSoftDelete"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From:             &User{},
			Where:             sq.And(userCol.Name, sq.LikeLeft("TestSoftDelete")),
			Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(1))
	}
	{
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: []sq.Insert{
				{userCol.ID, sq.UUID()},
				{userCol.Name, "TestSoftDelete"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From:             &User{},
			Where:             sq.And(userCol.Name, sq.LikeLeft("TestSoftDelete")),
			Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(2))
	}
	{
		result, err := testDB.SoftDelete(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.LikeLeft("TestSoftDelete")),
			Reviews: []string{"UPDATE `user` SET `deleted_at` = ? WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(2))
	}
	{
		_, err := testDB.SoftDelete(context.TODO(), sq.QB{
			From: &User{},
		})
		assert.EqualError(t, err, "Error 1064: You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'goclub/sql:(MAYBE_FORGET_WHERE)' at line 1")
	}
	{
		_, err := testDB.SoftDelete(context.TODO(), sq.QB{
			From: sq.Table("user",nil, nil),
		})
		assert.EqualError(t, err, "goclub/sql: SoftDelete(ctx, qb) qb.Form.SoftDeleteWhere().Query can not be empty string")
	}
}

func (suite TestDBSuite) TestSoftDeleteModel() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestSoftDeleteModel")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	newID := IDUser(sq.UUID())
	{
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: sq.Values{
				{userCol.ID, newID},
				{userCol.Name, "TestSoftDeleteModel"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	{
		count, err := testDB.Count(context.TODO(), sq.QB{
			From:             &User{},
			Where:             sq.And(userCol.Name, sq.LikeLeft("TestSoftDeleteModel")),
			Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
		})
		assert.NoError(t, err)
		assert.Equal(t, count, uint64(1))
	}
	// {
	// 	result, err := testDB.SoftDeleteModel(context.TODO(), &User{ID: newID,}, sq.QB{Reviews: []string{"UPDATE `user` SET `deleted_at` = ? WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?"}})
	// 	assert.NoError(t, err)
	// 	affected, err := result.RowsAffected()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, affected, int64(1))
	// }
	// {
	// 	count, err := testDB.Count(context.TODO(), sq.QB{
	// 		From:             &User{},
	// 		Where:             sq.And(userCol.Name, sq.LikeLeft("TestSoftDeleteModel")),
	// 		Reviews:          []string{"SELECT COUNT(*) FROM `user` WHERE `name` LIKE ? AND `deleted_at` IS NULL"},
	// 	})
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, count, uint64(0))
	// }
}

func (suite TestDBSuite) TestExecQB() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestExecQB")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		result, err := testDB.Insert(context.TODO(), sq.QB{
			From: &TableUser{},
			Insert: []sq.Insert{
				{userCol.ID, sq.UUID()},
				{userCol.Name, "TestExecQB"},
				{userCol.Age, 18},
			},
			Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)"},
		})
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, affected, int64(1))
	}
	result, err := testDB.ExecQB(context.TODO(), sq.QB{
		From: &User{},
		Update: sq.Set("name", "TestExecQB_changed"),
		Where: sq.And("name", sq.LikeLeft("TestExecQB")),
	}, sq.StatementUpdate)
	assert.NoError(t, err)
	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, affected, int64(1))
}


func (suite TestDBSuite) TestTransaction() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestTransaction")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			From:&User{},
			Where: sq.And("name", sq.Equal("TestTransaction_1")),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
	}
	{
		var execed bool
		 err := testDB.BeginTransaction(context.TODO(), sql.LevelReadCommitted,func(tx *sq.Transaction) sq.TxResult {
			execed = true
			_, err := tx.InsertModel(context.TODO(), &User{Name:"TestTransaction_1"}, sq.QB{Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`,`created_at`,`updated_at`) VALUES {#VALUES#}"}})
			assert.NoError(t, err)
			return tx.Commit()
		})
		assert.True(t, execed)
		assert.False(t, xerr.Is(err, sq.ErrTransactionIsRollback))
		assert.NoError(t, err)
	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			From:&User{},
			Where: sq.And("name", sq.Equal("TestTransaction_1")),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			From:&User{},
			Where: sq.And("name", sq.Equal("TestTransaction_2")),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
	}
	{
		var execed bool
		err := testDB.BeginTransaction(context.TODO(), sql.LevelReadCommitted,func(tx *sq.Transaction) sq.TxResult {
			execed = true
			_, err := tx.InsertModel(context.TODO(), &User{Name:"TestTransaction_2"},sq.QB{Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`,`created_at`,`updated_at`) VALUES (?,?,?,?,?)"}})
			assert.NoError(t, err)
			return tx.Rollback()
		})
		assert.True(t, execed)
		assert.True(t, xerr.Is(err, sq.ErrTransactionIsRollback))

	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			From:&User{},
			Where: sq.And("name", sq.Equal("TestTransaction_2")),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
	}

	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			From:&User{},
			Where: sq.And("name", sq.Equal("TestTransaction_3")),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
	}
	{
		var execed bool
		 err := testDB.BeginTransaction(context.TODO(), sql.LevelReadCommitted, func(tx *sq.Transaction) sq.TxResult {
			execed = true
			_, err := tx.InsertModel(context.TODO(), &User{Name:"TestTransaction_3"},sq.QB{Reviews: []string{"INSERT INTO `user` (`id`,`name`,`age`,`created_at`,`updated_at`) VALUES (?,?,?,?,?)"}})
			assert.NoError(t, err)
			return tx.RollbackWithError(xerr.New("custom error"))
		})
		assert.True(t, execed)
		assert.EqualError(t, err, "custom error")
	}
	{
		has, err := testDB.Has(context.TODO(), sq.QB{
			From:&User{},
			Where: sq.And("name", sq.Equal("TestTransaction_3")),
		})
		assert.NoError(t, err)
		assert.Equal(t, has, false)
	}
}



func (suite TestDBSuite) TestQueryRelation() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestQueryRelation")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: UserAddress{},
			Where: sq.And("address", sq.Like("TestQueryRelation_address")),
			Reviews: []string{"DELETE FROM `user_address` WHERE `address` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	newID := sq.UUID()
	{
		_, err := testDB.Insert(context.TODO(), sq.QB{
			From: &User{},
			Insert: []sq.Insert{
				{"id", newID},
				{"name", "TestQueryRelation"},
				{"age", 1},
			},
		})
		assert.NoError(t, err)
		_, err = testDB.Insert(context.TODO(), sq.QB{
			From: UserAddress{},
			Insert: []sq.Insert{
				{"user_id", newID},
				{"address", "TestQueryRelation_address"},
			},
		})
		assert.NoError(t, err)
	}
	{
		userWithAddress := UserWithAddress{}
		uaCol := userWithAddress.Column()
		has, err := testDB.QueryRelation(context.TODO(), &userWithAddress, sq.QB{
			Where: sq.And(uaCol.Name, sq.Equal("TestQueryRelation")),
			Reviews: []string{"SELECT `user`.`id` AS 'user.id', `user`.`name` AS 'user.name', `user`.`age` AS 'user.age', `user_address`.`address` AS 'user_address.address' FROM `user` LEFT JOIN `user_address` ON `user`.`id` = `user_address`.`user_id` WHERE `user`.`name` = ? AND `user`.`deleted_at` IS NULL AND `user_address`.`deleted_at` IS NULL LIMIT ?"},
		})
		assert.NoError(t, err)
		assert.Equal(t, has, true)
		assert.Equal(t, userWithAddress.Name, "TestQueryRelation")
		assert.Equal(t, userWithAddress.Age, 1)
		assert.Equal(t, userWithAddress.UserID, IDUser(newID))
		assert.Equal(t, userWithAddress.Address, sql.NullString{"TestQueryRelation_address", true})

	}
}


func (suite TestDBSuite) TestQueryRelationSlice() {
	t := suite.T()
	userCol := User{}.Column()
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: &User{},
			Where: sq.And(userCol.Name, sq.Like("TestQueryRelationSlice")),
			Reviews: []string{"DELETE FROM `user` WHERE `name` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	{
		_, err := testDB.ClearTestData(context.TODO(), sq.QB{
			From: UserAddress{},
			Where: sq.And("address", sq.Like("TestQueryRelationSlice_address")),
			Reviews: []string{"DELETE FROM `user_address` WHERE `address` LIKE ?"},
		})
		assert.NoError(t, err)
	}
	var idList []IDUser
	{
		for i:=0;i<2;i++{
			newID := sq.UUID()
			idList = append(idList, IDUser(newID))
			_, err := testDB.Insert(context.TODO(), sq.QB{
				From: &User{},
				Insert: []sq.Insert{
					{"id", newID},
					{"name", "TestQueryRelationSlice_" + strconv.Itoa(i)},
					{"age", i},
				},
			})
			assert.NoError(t, err)
			_, err = testDB.Insert(context.TODO(), sq.QB{
				From: UserAddress{},
				Insert: []sq.Insert{
					{"user_id", newID},
					{"address", "TestQueryRelationSlice_address_"  + strconv.Itoa(i)},
				},
			})
			assert.NoError(t, err)
		}
	}
	{
		var list  []UserWithAddress
		uaCol := UserWithAddress{}.Column()
		err := testDB.QueryRelationSlice(context.TODO(), &list, sq.QB{
			Where:    sq.And(uaCol.Name, sq.LikeLeft("TestQueryRelationSlice")),
			OrderBy:  []sq.OrderBy{{uaCol.Name, sq.ASC}},
			Reviews: []string{"SELECT `user`.`id` AS 'user.id', `user`.`name` AS 'user.name', `user`.`age` AS 'user.age', `user_address`.`address` AS 'user_address.address' FROM `user` LEFT JOIN `user_address` ON `user`.`id` = `user_address`.`user_id` WHERE `user`.`name` LIKE ? AND `user`.`deleted_at` IS NULL AND `user_address`.`deleted_at` IS NULL ORDER BY `user`.`name` ASC"},
		})
		assert.NoError(t, err)
		assert.Equal(t, len(list), 2)
		for i:=0;i<2;i++ {
			userWithAddress := list[i]
			assert.Equal(t, userWithAddress.Name, "TestQueryRelationSlice_" + strconv.Itoa(i))
			assert.Equal(t, userWithAddress.Age, i)
			assert.Equal(t, userWithAddress.UserID, idList[i])
			assert.Equal(t, userWithAddress.Address, sql.NullString{"TestQueryRelationSlice_address_" + strconv.Itoa(i), true})
		}
	}
}

func TestInsertNullInt(t *testing.T) {
	ctx := context.Background()
	{
		result, err := testDB.Insert(ctx, sq.QB{
			From: sq.Table("insert", nil, nil),
			Insert: []sq.Insert{
				{"age", sql.NullInt32{Valid:false, Int32: 1}},
			},
		}) ; assert.NoError(t, err)
		id, err := result.LastInsertId() ; assert.NoError(t, err)
		age := sql.NullInt32{}
		has, err := testDB.QueryRowScan(ctx, sq.QB{
			From: sq.Table("insert", nil, nil),
			Select:[]sq.Column{"age"},
			Where: sq.And("id", sq.Equal(id)),
		}, []interface{}{&age}) ; assert.NoError(t, err)
		assert.Equal(t,has, true)
		assert.Equal(t,age, sql.NullInt32{Valid:false})
	}
	{
		result, err := testDB.Insert(ctx, sq.QB{
			From: sq.Table("insert", nil, nil),
			Insert: []sq.Insert{
				{"age", sql.NullInt32{Valid:true, Int32: 1219}},
			},
		}) ; assert.NoError(t, err)
		id, err := result.LastInsertId() ; assert.NoError(t, err)
		age := sql.NullInt32{}
		has, err := testDB.QueryRowScan(ctx, sq.QB{
			From: sq.Table("insert", nil, nil),
			Select:[]sq.Column{"age"},
			Where: sq.And("id", sq.Equal(id)),
		}, []interface{}{&age}) ; assert.NoError(t, err)
		assert.Equal(t,has, true)
		assert.Equal(t,age, sql.NullInt32{Valid:true, Int32: 1219})
	}
}

func TestErrTransactionIsRollback(t *testing.T) {
	ctx := context.Background()
	err := testDB.BeginTransaction(ctx, sq.LevelReadCommitted, func(tx *sq.Transaction) sq.TxResult {
		return tx.Rollback()
	})
	assert.Equal(t,xerr.Is(err, sq.ErrTransactionIsRollback), true)
}

func TestLastQueryCost(t *testing.T) {
	ctx := context.Background()
	{
		err := testDB.BeginTransaction(ctx, sq.LevelReadCommitted, func(tx *sq.Transaction) sq.TxResult {
			cost, err := tx.LastQueryCost(ctx) ; if err != nil {
				return tx.RollbackWithError(err)
			}
			_=cost
			return tx.Commit()
		}) ; assert.NoError(t, err)
	}
	{
		cost, err := testDB.LastQueryCost(ctx)  ; assert.NoError(t, err)
		_=cost
	}
	testDB.PrintLastQueryCost(ctx)
}
