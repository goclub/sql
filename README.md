# goclub/sql
[![Go Reference](https://pkg.go.dev/badge/github.com/goclub/sql.svg)](https://pkg.go.dev/github.com/goclub/sql)


> goclub/sql 让你了解每一个函数执行的sql是什么，保证SQL的性能最大化的同时超越ORM的便捷。

## 指南

在 Go 中与 sql 交互一般会使用 [sqlx](https://github.com/jmoiron/sqlx) 或 [gorm](http://gorm.io/) [xorm](https://xorm.io/zh/)

`sqlx` 偏底层，是对 `database/sql` 的封装，主要提供了基于结构体标签 `db:"name""`将查询结果解析为结构体的功能。而GORM XORM 功能则更丰富。

直接使用 `sqlx` 频繁的手写 sql 非常繁琐且容易出错。（database/sql 的接口设计的不是很友好）。

GORM XORM 存在 ORM 都有的特点，使用者容易使用 ORM 运行一些性能不高的 SQL。虽然合理使用也可以写出高效SQL，但使用者在使用 ORM 的时候容易忽略最终运行的SQL是什么。

[goclub/sql](https://github.com/goclub/sql) 提供介于手写 sql 和 ORM 之间的使用体验。


## Open

> 连接数据库

goclub/sql 与 database/sql 连接方式相同，只是多返回了 dbClose 函数。 `dbClose` 等同于 `db.Close`

[Open](./example/internal/connect/main.go?embed)


## ExecMigrate

> 通过迁移代码创建表结构 

[创建用户迁移文件](./example/internal/migrate/migrate/20201004160444_user.go?embed)

[ExecMigrate](./example/internal/migrate/main.go?embed)

## 定义Model

通过表单创建 Model: [goclub.run](https://goclub.run/?k=model)

## Insert

> 使用 Insert 插入数据

[Insert](./example/internal/insert/main.go?embed)

## InsertModel

> 基于 Model 插入数据

大部分场景下使用 `db.Insert` 插入数据有点繁琐。基于 `sq.Model` 使用 `db.InsertModel`操作数据会方便很多。

[InsertModel](./example/internal/insert_model/main.go?embed)


## Update

> 使用 Update 更新数据

[Update](./example/internal/update/main.go?embed)

> goclub/sql 故意没有提供 UpdateModel 方法, 使用 `db.Update(ctx, sq.QB{...})` 精准的更新数据

## Query 

> 使用 Query 查询单条数据
> 使用 QuerySlice 查询多条数据

[Query](./example/internal/query/main.go?embed)

> goclub/sql 故意没有提供 QueryModel 方法, 使用 `db.Query(ctx, &user, sq.QB{ Where: sq.And(col.ID, sq.Equal(userID)) })` 可以查询 Model
 
## SoftDelete HardDelete

> 使用SoftDelete 或者 HardDelete 删除数据 

[delete](./example/internal/delete/main.go?embed)

## Relation

[relation](./example/internal/relation/main.go?embed)

## 致谢

> 感谢 [jetbrains](https://jb.gg/OpenSource) 提供 Goland 开源授权

