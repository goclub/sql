package sq_test

import sq "github.com/goclub/sql"

type Migrate struct {

}
func (Migrate) Migrate20201004160444CreateUserTable(mi sq.Migrate) {

	mi.CreateTable(sq.CreateTableQB{
		TableName: "user",
		PrimaryKey: []string{"id"},
		Fields: append([]sq.MigrateField{
			mi.Field("id").Char(36).DefaultString(""),
			mi.Field("name").Varchar(255).DefaultString(""),
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

