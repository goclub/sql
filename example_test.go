package sq_test

import (
	"context"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goclub/sql"
	"log"
	"testing"
)

func TestExample(t *testing.T) {
	ExampleDB_QueryRowScan()
	ExampleDB_QueryRowScanMultiColumn()
	ExampleDB_QueryRowStructScan()
	ExampleDB_Count()
	ExampleCreateModel()
	ExampleMultiCreateModel()
	ExampleDB_One()
	ExampleDB_ModelList()
	ExampleDB_UpdateModel()
	ExampleSoftDeleteModel()
	ExampleUpdate()
	ExampleRelation()
	ExampleRelationList()
}
var exampleDB *sq.DB
func init () {
	db, dbClose, err := sq.Open("mysql", sq.DataSourceName{
		DriverName: "mysql",
		User: "root",
		Password:"somepass",
		Host: "127.0.0.1",
		Port:"3306",
		DB: "goclub_sql",
	}.String()) ; if err != nil {
		panic(err)
	}
	exampleDB = db
	_=dbClose // init 场景下不需要 close，应该在 main 执行完毕后 close
	err = exampleDB.Core.Ping() ; if err != nil {
		panic(err)
	}
}


type IDUser string
type User struct {
	ID IDUser `db:"id"`
	Name string `db:"name"`
	Age int `db:"age"`
}
func (User) TableName() string {return "user"}
func (u *User) BeforeCreate() {
	if len(u.ID) == 0 { u.ID = IDUser(sq.UUID()) }
}
func (User) Column () (col struct{
	ID sq.Column
	Name sq.Column
	Age sq.Column
}) {
	col.ID = "id"
	col.Name = "name"
	col.Age = "age"
	return
}

type UserWithAddress struct {
	UserID IDUser `db:"user.id"`
	Name string `db:"user.name"`
	Age int `db:"user.age"`
	Address string `db:"user_address.address"`
}
func (UserWithAddress) FormTable() string {return "user"}
func (*UserWithAddress) RelationJoin() []sq.Join {
	return []sq.Join{
		{
			Type: 	  	   sq.LeftJoin,
			TableName:	   "user_address",
			On:[]sq.Column{"user.id", "user_address.user_id"},
		},
	}
}
func (UserWithAddress) Column () (col struct{
	UserID sq.Column
	Name sq.Column
	Age sq.Column
	Address sq.Column
}) {
	col.UserID = "user.id"
	col.Name = "user.name"
	col.Age = "user.age"
	col.Address = "user_address.user_id"
	return
}
// 查询单行单列数据
func ExampleDB_QueryRowScan() {
	log.Print("ExampleDB_QueryRowScan")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userCol := User{}.Column()
	var name string
	qb := sq.QB{
		Table: User{}.TableName(),
		Select: []sq.Column{userCol.Name},
		Where: sq.
			And(userCol.ID, sq.Equal(1)),
	}.Check("SELECT `name` FROM `user` WHERE `id` = ? AND deleted_at IS NULL LIMIT ?")
	hasName, err := exampleDB.QueryRowScan(ctx, qb, &name) ; if err != nil {
		panic(err)
	}
	log.Print(name, hasName)
}
// 查询单行多列数据
func ExampleDB_QueryRowScanMultiColumn() {
	log.Print("ExampleDB_QueryRowScanMultiColumn")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userCol := User{}.Column()
	var name string
	var age int
	qb := sq.QB{
		Table: User{}.TableName(),
		Select: []sq.Column{userCol.Name, userCol.Age},
		Where: sq.
			And(userCol.ID, sq.Equal(1)),
	}.Check("SELECT `name`,`age` FROM `user` WHERE `id` = ? AND deleted_at IS NULL LIMIT ?")
	hasName, err := exampleDB.QueryRowScan(ctx, qb, &name,&age) ; if err != nil {
		panic(err)
	}
	log.Print(name, hasName)
}
// 查询单行多列数据(扫描到结构体)
func ExampleDB_QueryRowStructScan() {
	log.Print("ExampleDB_QueryRowStructScan")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userNameAge := struct {
		Name string `db:"name"`
		Age int `db:"age"`
	}{}
	userCol := User{}.Column()
	qb := sq.QB{
		Table: User{}.TableName(),
		Where: sq.
			And(userCol.Name, sq.Equal("nimo")).
			And(userCol.Age, sq.Equal(18)),
	}.Check("SELECT `name`,`age` FROM `user` WHERE `name` = ? AND `age` = ? AND deleted_at IS NULL LIMIT ?")
	hasUser, err := exampleDB.QueryRowStructScan(ctx, &userNameAge, qb) ; if err != nil {
		panic(err)
	}
	log.Print(userNameAge, hasUser)
}

// 基于 Model 查询单行数据 （可省略 qb.Table）
func ExampleDB_One() {
	log.Print("ExampleDB_One")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	user := User{}
	userCol := user.Column()
	qb := sq.QB{
		Where: sq.
			And(userCol.Name, sq.Equal("nimo")).
			And(userCol.Age, sq.Equal(18)),
	}.Check("SELECT `id`,`name`,`age`,`deleted_at` FROM `user` WHERE `name` = ? AND `age` = ? AND deleted_at IS NULL LIMIT ?")
	hasUser, err := exampleDB.Model(ctx, &user, qb) ; if err != nil {
		panic(err)
	}
	log.Print(user, hasUser)
}
//  基于 Model 查询 count
func ExampleDB_Count() {
	log.Print("ExampleDB_Count")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	count, err := exampleDB.Count(ctx, sq.QB{
		Table: User{}.TableName(),
		Where: sq.And(User{}.Column().Age, sq.GtInt(18)),
	}) ; if err != nil {
		panic(err)
	}
	log.Print(count)
}
// 基于 Model 查询多行数据
func ExampleDB_ModelList() {
	log.Print("ExampleDB_ModelList")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	var userList []User
	userCol := User{}.Column()
	qb := sq.QB{
		Where: sq.
			And(userCol.Age, sq.GtInt(18)),
	}.Check("SELECT `id`,`name`,`age`,`deleted_at` FROM `user` WHERE `age` > ? AND deleted_at IS NULL")
	err := exampleDB.ModelList(ctx, &userList, qb) ; if err != nil {
		panic(err)
	}
	log.Print(userList)
}
func someUser () (user User) {
	userCol := user.Column()
	{
		oneQB := sq.QB{
			Where: sq.And(userCol.Name, sq.Equal("update1"),),
		}.Check("SELECT `id`, `name`, `age` FROM `user` WHERE `name` = ? AND `deleted_at` IS NULL LIMIT ?")
		hasUser, err := exampleDB.Model(context.TODO(), &user, oneQB) ; if err != nil {
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
	updateCheckSQL := "UPDATE `user` SET `name` = ? WHERE `id` = ? AND `deleted_at` IS NULL"
	err := exampleDB.UpdateModel(ctx, &user, sq.UpdateColumn{
		userCol.Name: "newUpdate",
	}, updateCheckSQL) ; if err != nil {
		panic(err)
	}
	log.Print(user.Name) // newUpdate ( db.UpdateModel() 会自动给 user 对应字段赋值)
	// 在已知主键字段的情况下可以不读取 Model
	someID := IDUser("290f187c-3de0-11eb-b378-0242ac130002")
	err = exampleDB.UpdateModel(ctx, &User{
		ID: someID,
	}, sq.UpdateColumn{userCol.Name: "newUpdate",}, updateCheckSQL) ; if err != nil {
		panic(err)
	}
}

func ExampleUpdate() {
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userCol := User{}.Column()
	err := exampleDB.Update(ctx, sq.QB{
		Table:  User{}.TableName(),
		Where:  sq.And(userCol.Name, sq.Equal("multiUpdate")),
		Update: sq.UpdateColumn{
			userCol.Age: 28,
		},
	}.Check("UPDATE `user` SET `age` = ? WHERE `name` = ? AND `deleted_at` IS NULL"))
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
func ExampleMultiCreateModel() {
	log.Print("ExampleMultiCreateModel")
	ctx := context.TODO() // 一般由 http.Request{}.Context() 获取
	userList := []User{
		{Name:"a", Age:1},
		{Name:"b", Age:2},
	}
	checkInsertSQL := "INSERT INTO `user` (`id`, `name`, `age`) VALUES (?, ?, ?)"
	err := exampleDB.MultiCreateModel(ctx, &userList, checkInsertSQL) ; if err != nil {
		panic(err)
	}
}
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
	err := exampleDB.Relation(ctx, &userWithAddress, sq.QB{Where: sq.And(userWithAddressCol.UserID, sq.Equal("290f187c-3de0-11eb-b378-0242ac130002"))}, checkSQL) ; if err != nil {
		panic(err)
	}
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
	err := exampleDB.RelationList(ctx, &userWithAddressList, sq.QB{Where: sq.And(userWithAddressCol.Age, sq.GtInt(18))}, checkSQL) ; if err != nil {
		panic(err)
	}
}