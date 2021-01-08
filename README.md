# sql
[![Go Reference](https://pkg.go.dev/badge/github.com/goclub/sql.svg)](https://pkg.go.dev/github.com/goclub/sql)

![](./cat.png)

> goclub/sql 让你了解每一个函数执行的sql是什么，保证SQL的性能最大化的同时超越ORM的便捷。

## 指南

在 Go 中与 sql 交互一般会使用 [sqlx](https://github.com/jmoiron/sqlx) 或 [gorm](http://gorm.io/) [xorm](https://xorm.io/zh/)

`sqlx` 偏底层，是对 `database/sql` 的封装，主要提供了基于结构体标签 `db:"name""`将查询结果解析为结构体的功能。而GORM XORM 功能则更丰富。

直接使用 `sqlx` 频繁的手写 sql 非常繁琐且容易出错。（database/sql 的接口设计的不是很友好）。

GORM XORM 存在 ORM 都有的特点，使用者容易使用 ORM 运行一些性能不高的 SQL。虽然合理使用也可以写出高效SQL，但使用者在使用 ORM 的时候容易忽略最终运行的SQL是什么。

[goclub/sql](https://github.com/goclub/sql) 提供介于手写 sql 和 ORM 之间的使用体验。


**查询单行单列**

```sql
SELECT `name` FROM `user` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT ?
```
```go
var name string
var hasName bool
qb := sq.QB{
  Table: TableUser{},
  Select: []sq.Column{"name"},
  Where: sq.And("id", sq.Equal(1)),
}
hasName := exampleDB.QueryRowScan(ctx,qb,&name)
```

**查询单行多列数据(扫描到结构体)**

```sql
SELECT `name`, `age` FROM `user` WHERE `name` = ? AND `age` = ? AND `deleted_at` IS NULL LIMIT ?
```

```go
type UserNameAge struct {
    Name string `db:"name"`
    Age int `db:"age"`
    TableUser // 定义表信息 https://github.com/goclub/sql/blob/main/example_user_test.go
}
var userNameAge UserNameAge
var hasUser bool
qb := sq.QB{
    Table: UserNameAge{},
    Where: sq.
        And(userCol.Name, sq.Equal("nimo")).
        And(userCol.Age, sq.Equal(18)),
}
hasUser, err := exampleDB.QueryRowStructScan(ctx, &userNameAge, qb) ; if err != nil {
    panic(err)
}
```

**查询多行单列数据**

```sql
SELECT `id` FROM `user` WHERE `deleted_at` IS NULL
```

```go
qb := sq.QB{
    Table: TableUser{},
    Select: []sq.Column{"id"},
}
var userIDList []string
err := exampleDB.SelectScan(ctx, qb, sq.ScanStrings(*userIDList)) ; if err != nil {
    panic(err)
}
```

**查询多行多列数据解析结构体切片**

```sql
SELECT `name`, `age` FROM `user` WHERE `age` = ? AND `deleted_at` IS NULL
```

```go
type UserNameAge struct {
    Name string `db:"name"`
    Age int `db:"age"`
    TableUser // https://github.com/goclub/sql/blob/main/example_user_test.go
}
userNameAgeList := []UserNameAge{}
qb := sq.QB{
    Table: UserNameAge{},
    Where: sq.And("age", sq.Equal(18)),
}
err := exampleDB.Select(ctx, &userNameAgeList, qb) ; if err != nil {
    panic(err)
}
```


**基于 Model 查询单行数据**

```sql
SELECT `id`, `name`, `age`, `created_at`, `updated_at` FROM `user` WHERE `name` = ? AND `age` = ? LIMIT ?
```

```go
user := User{} 
var hasUser bool
qb := sq.QB{
    Where: sq.
        And("name", sq.Equal("nimo")).
        And("age", sq.Equal(18)),
}
hasUser, err := exampleDB.QueryModel(ctx, &user, qb) ; if err != nil {
    panic(err)
}
```

## 教程

> 推荐不了解 database/sql 的使用者阅读： [Go SQL 数据库教程
](https://learnku.com/docs/go-database-sql/overview/9474)

## 准备工作

[定义结构体](https://github.com/goclub/sql/blob/main/example_user_test.go)

### QueryRowScan
 
[查询单行多列数据](https://pkg.go.dev/github.com/goclub/sql/#example-DB.QueryRowScan)

### QueryRowStructScan

[查询单行数据并解析到结构体](https://pkg.go.dev/github.com/goclub/sql/#example-DB.QueryRowStructScan)

### SelectScan

[查询多行单列数据](https://pkg.go.dev/github.com/goclub/sql/#example-DB.SelectScan)

### Select
 
[查询多行多列数据解析结构体切片](https://pkg.go.dev/github.com/goclub/sql/#example-DB.Select)

### QueryModel

[基于 Model 查询单行数据](https://pkg.go.dev/github.com/goclub/sql/#example-DB.QueryModel)

若无需查询全部字段可考虑使用 QueryRowStructScan 查询

### QueryModelList

[基于 Model 查询多行数据](https://pkg.go.dev/github.com/goclub/sql/#example-DB.QueryModelList)

### Count

[计数查询 count(*)](https://pkg.go.dev/github.com/goclub/sql/#example-DB.Count)
