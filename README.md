# goclub/sql
[![Go Reference](https://pkg.go.dev/badge/github.com/goclub/sql.svg)](https://pkg.go.dev/github.com/goclub/sql)

![](./cat.png)

> goclub/sql 让你了解每一个函数执行的sql是什么，保证SQL的性能最大化的同时超越ORM的便捷。

## 指南

在 Go 中与 sql 交互一般会使用 [sqlx](https://github.com/jmoiron/sqlx) 或 [gorm](http://gorm.io/) [xorm](https://xorm.io/zh/)

`sqlx` 偏底层，是对 `database/sql` 的封装，主要提供了基于结构体标签 `db:"name""`将查询结果解析为结构体的功能。而GORM XORM 功能则更丰富。

直接使用 `sqlx` 频繁的手写 sql 非常繁琐且容易出错。（database/sql 的接口设计的不是很友好）。

GORM XORM 存在 ORM 都有的特点，使用者容易使用 ORM 运行一些性能不高的 SQL。虽然合理使用也可以写出高效SQL，但使用者在使用 ORM 的时候容易忽略最终运行的SQL是什么。

[goclub/sql](https://github.com/goclub/sql) 提供介于手写 sql 和 ORM 之间的使用体验。


## 连接数据库

goclub/sql 与 database/sql 连接方式相同，只是多返回了 dbClose 函数。 `dbClose` 等同于 `db.Close`

[connect | embed](./exmaple/internal/connect/main.go)


## 通过迁移创建表

> 新建并编辑迁移命令 

[创建用户迁移文件 | embed](./example/internal/migrate/actions/20201004160444_create_user_table.go)

> 创建入口文件

[入口 | embed](./example/internal/migrate/main.go)

## 准备表


[可视化配置 model](https://tools.goclub.vip/?tag=model)

## 插入数据

> INSERT INTO `user` (`id`,`name`,`age`) VALUES (?,?,?)

[insert | embed](./example/internal/insert/main.go)

## 基于 Model 插入数据

大部分场景下使用 `db.Insert` 插入数据有点繁琐。基于 `sq.Model` 使用 `db.InsertModel`操作数据会方便很多。

