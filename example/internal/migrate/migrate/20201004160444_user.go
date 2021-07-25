package migrate

import sq "github.com/goclub/sql"

type Migrate struct {

}

func (Migrate) Migrate20201004160444CreateUserTable(mi sq.Migrate) {

	mi.CreateTable(sq.CreateTableQB{
		TableName: "user",
		PrimaryKey: []string{"id"},
		Fields: append([]sq.MigrateField{
			mi.Field("id").Type("bigint",20).AutoIncrement(),
			mi.Field("mobile").Char(11).DefaultString(""),
			mi.Field("name").Char(20).DefaultString(""),
			mi.Field("china_id_card_no").Char(18).Null(),
		}, mi.CUDTimestamp()...),
		Key: map[string][]string{
			"mobile": {"mobile"},
		},
		UniqueKey: map[string][]string{
			"china_id_card_no": []string{"china_id_card_no"},
		},
		Engine: mi.Engine().InnoDB,
		Charset: mi.Charset().Utf8mb4,
		Collate: mi.Utf8mb4_unicode_ci(),
	})
	mi.CreateTable(sq.CreateTableQB{
		TableName: "user_address",
		PrimaryKey: []string{"user_id"},
		Fields: append([]sq.MigrateField{
			mi.Field("user_id").Type("bigint", 20),
			mi.Field("address").Varchar(255).DefaultString(""),
		}, mi.CUDTimestamp()...),
		Key: map[string][]string{},
		UniqueKey: map[string][]string{},
		Engine: mi.Engine().InnoDB,
		Charset: mi.Charset().Utf8mb4,
		Collate: mi.Utf8mb4_unicode_ci(),
	})
}
