package migrateActions

import sq "github.com/goclub/sql"

func (Migrate) Migrate20201004160444CreateUserTable(mi sq.Migrate) {
	// 还可以使用 mi.Exec
	// mi.Exec("CREATE TABLE `demo` ( `id` int(11) unsigned NOT NULL AUTO_INCREMENT, PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	// 或者直接使用 mi.DB
	// mi.DB
	mi.CreateTable(sq.CreateTableQB{
		TableName: "user",
		PrimaryKey: []string{"id"},
		Fields: append([]sq.MigrateField{
			mi.Field("id").Char(36).DefaultString(""),
			mi.Field("name").Varchar(20).DefaultString(""),
			mi.Field("age").Int(11).DefaultInt(0),
		}, mi.CUDTimestamp()...),
		Key: map[string][]string{
			"name": {"name"},
		},
		Engine: mi.Engine().InnoDB,
		Charset: mi.Charset().Utf8mb4,
		Collate: mi.Utf8mb4_unicode_ci(),
	})
	mi.CreateTable(sq.CreateTableQB{
		TableName: "user_address",
		PrimaryKey: []string{"user_id"},
		Fields: append([]sq.MigrateField{
			mi.Field("user_id").Char(36).DefaultString(""),
			mi.Field("address").Varchar(255).DefaultString(""),
		}, mi.CUDTimestamp()...),
		Key: map[string][]string{

		},
		Engine: mi.Engine().InnoDB,
		Charset: mi.Charset().Utf8mb4,
		Collate: mi.Utf8mb4_unicode_ci(),
	})
}
