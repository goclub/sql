package sq_test

import (
	"context"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goclub/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"testing"
)


func TestExample(t *testing.T) {
	// ExampleDB_QueryRowScan()
	// ExampleDB_QueryRowStructScan()
	ExampleDB_SelectScan()
	// ExampleDB_Select()
	// ExampleDB_QueryModel()
	// ExampleDB_Count()
	// ExampleDB_ModelList()

}
var exampleDB *sq.DB
func init () {
	db, dbClose, err := sq.Open("mysql", sq.DataSourceName{
		DriverName: "mysql",
		User: "root",
		Password:"somepass",
		Host: "127.0.0.1",
		Port:"3306",
		DB: "example_goclub_sql",
		Query: map[string]string{
			"parseTime": "true",
		},
	}.String()) ; if err != nil {
		panic(err)
	}
	exampleDB = db
	_=dbClose // init 场景下不需要 close，应该在 main 执行完毕后 close
	err = exampleDB.Core.Ping() ; if err != nil {
		panic(err)
	}
}

// 查询单行多列数据
// sq.QB 是 goclub/sql 的核心功能。 QB = query builder 用于生成 SQL。
//
// &name 用于接收查询结果
//
// hasName 当没有查询到数据时 hasName = false ，否则为 true
//
// 通过 qb.Debug = true 可以在执行 SQL 时打印 sql语句和占位符对应的值
func ExampleDB_QueryRowScan() {
	log.Print("ExampleDB_QueryRowScan")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	{
		log.Print("查询单行单列")
		var name string
		qb := sq.QB{
			Debug: true, // Debug 时候会通过 log.Print 打印执行的 SQL
			Table: TableUser{},
			Select: []sq.Column{"name"},
			Where: sq.
				And("id", sq.Equal(1)),
		}
		// SELECT `name` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?
		hasName, err := exampleDB.QueryRowScan(ctx, qb, &name) ; if err != nil {
			panic(err)
		}
		log.Print(" name:", name, " hasName:", hasName)
	}
	{
		log.Print("查询单行多列")
		var name string
		var age int
		qb := sq.QB{
			Debug: true,
			Table: TableUser{},
			Select: []sq.Column{"name","age"},
			Where: sq.
				And("id", sq.Equal(1)),
		}
		// SELECT `name`, `age` FROM `user` WHERE `id` = ? AND `is_deleted` = 0 LIMIT ?
		hasUser, err := exampleDB.QueryRowScan(ctx, qb, &name, &age) ; if err != nil {
			panic(err)
		}
		log.Print(" name:", name, " age", age, " hasUser:", hasUser)
	}
}
// 查询单行多列数据(扫描到结构体)
func ExampleDB_QueryRowStructScan() {
	log.Print("ExampleDB_QueryRowStructScan")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	// 定义查询结果对应的结构体，并组合  TableUser 以提供表名和软删信息
	type UserNameAge struct {
		Name string `db:"name"`
		Age int `db:"age"`
		TableUser // https://github.com/goclub/sql/blob/main/example_user_test.go
	}
	userNameAge := UserNameAge{}
	userCol := User{}.Column()
	qb := sq.QB{
		Debug: true,
		Table: UserNameAge{},
		// Select 为空时候会根据 qb.Table (UserNameAge{})  结构体每个字段的 `db:"xxx"`作为 Select 参数
		Where: sq.
			And(userCol.Name, sq.Equal("nimo")).
			And(userCol.Age, sq.Equal(18)),
	}
	// SELECT `name`, `age` FROM `user` WHERE `name` = ? AND `age` = ? AND `deleted_at` IS NULL LIMIT ?
	hasUser, err := exampleDB.QueryRowStructScan(ctx, &userNameAge, qb) ; if err != nil {
		panic(err)
	}
	log.Print("userNameAge.Name:",userNameAge.Name, " userNameAge.Age:",userNameAge.Age, " hasUser:", hasUser)
}
// 查询多行单列数据
func ExampleDB_SelectScan() {
	log.Print("ExampleDB_SelectScan")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	qb := sq.QB{
		Debug: true,
		Table: TableUser{},
		Select: []sq.Column{User{}.Column().ID},
	}
	var userIDList []IDUser
	// SELECT `id` FROM `user` WHERE `deleted_at` IS NULL
	err := exampleDB.SelectScan(ctx, qb, func(rows *sqlx.Rows) error {
		var userID IDUser
		err := rows.Scan(&userID) ; if err != nil {
			return err
		}
		userIDList = append(userIDList, userID)
		return nil
	}) ; if err != nil {
		panic(err)
	}
	log.Print("userIDList:", userIDList)
}
// 查询多行多列数据解析结构体切片
func ExampleDB_Select() {
	log.Print("ExampleDB_Select")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	// 定义查询结果对应的结构体，并组合  TableUser 以提供表名和软删信息
	type UserNameAge struct {
		Name string `db:"name"`
		Age int `db:"age"`
		TableUser // https://github.com/goclub/sql/blob/main/example_user_test.go
	}
	userNameAgeList := []UserNameAge{}
	userCol := User{}.Column()
	qb := sq.QB{
		Debug: true,
		Table: UserNameAge{},
		Where: sq.
			And(userCol.Age, sq.Equal(18)),
	}
	// SELECT `name`, `age` FROM `user` WHERE `age` = ? AND `deleted_at` IS NULL
	err := exampleDB.Select(ctx, &userNameAgeList, qb) ; if err != nil {
		panic(err)
	}
	log.Print("userNameAgeList:", userNameAgeList)
}

// 基于 Model 查询单行数据 （可不传递 qb.Table）
func ExampleDB_QueryModel() {
	log.Print("ExampleDB_Model")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	user := User{}
	userCol := user.Column()
	qb := sq.QB{
		Debug: true,
		Where: sq.
			And(userCol.Name, sq.Equal("nimo")).
			And(userCol.Age, sq.Equal(18)),
	}
	// SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` = ? AND `age` = ? LIMIT ?
	hasUser, err := exampleDB.QueryModel(ctx, &user, qb) ; if err != nil {
		panic(err)
	}
	log.Printf("user: %+v\r\n hasUser: %v", user, hasUser)
}
// 基于 Model 查询多行数据
func ExampleDB_ModelList() {
	log.Print("ExampleDB_ModelList")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	var userList []User
	userCol := User{}.Column()
	qb := sq.QB{
		Debug: true,
		Where: sq.
			And(userCol.Age, sq.GtInt(10)),
	}
	// SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL
	err := exampleDB.QueryModelList(ctx, &userList, qb) ; if err != nil {
		panic(err)
	}
	log.Print(userList)
}
//  基于 Model 查询 count
func ExampleDB_Count() {
	log.Print("ExampleDB_Count")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	// SELECT COUNT(*) FROM `user` WHERE `age` = ? AND `deleted_at` IS NULL
	count, err := exampleDB.Count(ctx, sq.QB{
		Debug: true,
		Table: User{},
		Where: sq.And(User{}.Column().Age, sq.Equal(18)),
	}) ; if err != nil {
		panic(err)
	}
	log.Print("count: ", count)
}
func someUser () (user User) {
	userCol := user.Column()
	{
		hasUser, err := exampleDB.QueryModel(context.TODO(), &user, sq.QB{
			Where: sq.And(userCol.Name, sq.Equal("update1"),),
		}) ; if err != nil {
		panic(err)
	}
		if hasUser == false {
			panic(errors.New(`example data not found user{name: "update1"}`))
		}
	}
	return
}
func ExampleDB_UpdateModel() {
	log.Print("ExampleDB_UpdateModel")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	var user User
	user = someUser()
	userCol := user.Column()
	// Update() 会优先以 `id` = user.ID 作为 WHERE 条件
	// 若 user 不存在 user.ID 则以包含结构体标签 `sq:"PRI"` 的字段作为 WHERE 条件
	// 存在多个 `sq:"PRI"`则以多个条件查询
	checkSQL := "UPDATE `user` SET `name` = ? WHERE `id` = ? AND `deleted_at` IS NULL"
	err := exampleDB.UpdateModel(ctx, &user, []sq.Data{
		{userCol.Name, "newUpdate"},
	}, checkSQL) ; if err != nil {
		panic(err)
	}
	log.Print(user.Name) // newUpdate ( db.UpdateModel() 会自动给 user 对应字段赋值)
	// 在已知主键字段的情况下可以不读取 Model
	someID := IDUser("290f187c-3de0-11eb-b378-0242ac130002")
	err = exampleDB.UpdateModel(ctx, &User{
		ID: someID,
	}, []sq.Data{{userCol.Name, "newUpdate",},}, checkSQL) ; if err != nil {
		panic(err)
	}
}

func ExampleUpdate() {
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userCol := User{}.Column()
	checkSQL := "UPDATE `user` SET `age` = ? WHERE `name` = ? AND `deleted_at` IS NULL"
	err := exampleDB.Update(ctx, sq.QB{
		Table: User{},
		Where:  sq.And(userCol.Name, sq.Equal("multiUpdate")),
		Update: []sq.Data{
			{userCol.Age, 28,},
		},
	}.Check(checkSQL))
	if err != nil {
		panic(err)
	}
}
func ExampleCreateModel() {
	log.Print("ExampleCreateModel")
	ctx := context.TODO()
	var user User
	checkInsertSQL := "INSERT INTO `user` (`id`, `name`, `age`) VALUES (?, ?, ?)"
	err := exampleDB.CreateModel(ctx, &user, checkInsertSQL) ; if err != nil {
		panic(err)
	}
}
// func ExampleMultiCreateModel() {
// 	log.Print("ExampleMultiCreateModel")
// 	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
// 	userList := []User{
// 		{Name:"a", Age:1},
// 		{Name:"b", Age:2},
// 	}
// 	checkInsertSQL := "INSERT INTO `user` (`id`, `name`, `age`) VALUES (?, ?, ?)"
// 	err := exampleDB.MultiCreateModel(ctx, &userList, checkInsertSQL) ; if err != nil {
// 		panic(err)
// 	}
// }
func ExampleSoftDeleteModel() {
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	user := someUser()
	softDeletedCheckSQL := "UPDATE `user` SET `deleted_at` = NULL"
	err := exampleDB.SoftDeleteModel(ctx, &user, softDeletedCheckSQL) ; if err != nil {
		panic(err)
	}
}
func ExampleRelation() {
	log.Print("ExampleRelation")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userWithAddress := UserWithAddress{}
	userWithAddressCol := userWithAddress.Column()
	checkSQL := "SELECT `user.id`, `user.name`, `user.age`, `user_address.address` " +
		"FROM `user` " +
		"LEFT JOIN `user_address` " +
		"ON `user.id` = `user_address.id " +
		"WHERE `user.id` = ? " +
		"AND `user.deleted_at` IS NULL " +
		"AND `user_address.deleted_at` IS NULL" +
		"LIMIT ?"
	has, err := exampleDB.QueryRelation(ctx, &userWithAddress, sq.QB{Where: sq.And(userWithAddressCol.UserID, sq.Equal("290f187c-3de0-11eb-b378-0242ac130002"))}, checkSQL) ; if err != nil {
		panic(err)
	}
	log.Print("has", has)
}

func ExampleRelationList() {
	log.Print("ExampleRelationList")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	var userWithAddressList []UserWithAddress
	userWithAddressCol := UserWithAddress{}.Column()
	checkSQL := "SELECT `user.id`, `user.name`, `user.age`, `user_address.address` " +
		"FROM `user` " +
		"LEFT JOIN `user_address` " +
		"ON `user.id` = `user_address.id " +
		"WHERE `user.age` > ? " +
		"AND `user.deleted_at` IS NULL " +
		"AND `user_address.deleted_at` IS NULL"
	err := exampleDB.QueryRelationList(ctx, &userWithAddressList, sq.QB{Where: sq.And(userWithAddressCol.Age, sq.GtInt(18))}, checkSQL) ; if err != nil {
		panic(err)
	}
}
func ExamplePaging() {
	log.Print("ExamplePaging")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	var userList []User
	userCol := User{}.Column()
	baseQB := sq.QB{
		Table: User{},
		Where: sq.And(userCol.Age, sq.GtInt(18)),
	}
	page := 1
	perPage := 10
	checkSQL := "SELECT `id`, `name`, `age` FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL LIMIT ? OFFSET ?"
	pagingQB :=  baseQB.Paging(page, perPage)
	err := exampleDB.QueryModelList(ctx, &userList, pagingQB.Check(checkSQL)) ; if err != nil {
		panic(err)
	}
	log.Print(userList)
	var count int
	checkCountSQL := "SELECT COUNT(*) FROM `user` WHERE `age` > ? AND `deleted_at` IS NULL"
	count, err = exampleDB.Count(ctx, baseQB.Check(checkCountSQL)) ; if err != nil {
		panic(err)
	}
	log.Print(count)
}