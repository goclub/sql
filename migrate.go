package sq

import (
	"context"
	xerr "github.com/goclub/error"
	"log"
	"reflect"
	"strings"
)


const createMigratestringQueueL = `
CREATE TABLE IF NOT EXISTS goclub_sql_migrations (
  id int(10) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 
`

func ExecMigrate(db *Database, ptr interface{}) (err error) {
	ctx := context.Background()
	table := Table("goclub_sql_migrations", nil, nil)
	rPtrValue := reflect.ValueOf(ptr)
	if rPtrValue.Kind() != reflect.Ptr {
		panic(xerr.New("ExecMigrate(db, ptr) ptr must be pointer"))
	}
	rValue := rPtrValue.Elem()
	rType := rValue.Type()
	if _, err = db.Exec(ctx, createMigratestringQueueL, nil); err != nil {
	    return
	}
	methodNames := []string{}
	for i:=0;i<rType.NumMethod();i++ {
		method := rType.Method(i)
		if strings.HasPrefix(method.Name, "Migrate") {
			methodNames = append(methodNames, method.Name)
		}
	}
	for _, methodName := range methodNames {
		var has bool
		if has, err = db.Has(ctx, table, QB{
 			Where: And("name", Equal(methodName)),
 		}); err != nil {
		    return
		}
		if has  {
			continue
		}
		log.Print("[goclub_sql migrate]exec: " +methodName)
		out := rValue.MethodByName(methodName).Call([]reflect.Value{})
		if len(out) != 1 {
			return xerr.New(methodName + "() must return error or nil")
		}
		errOrNil := out[0].Interface()
		if errOrNil != nil {
			return errOrNil.(error)
		}
		if _, err = db.Insert(ctx, QB{
 			From: table,
 			Insert: Values{
 				{"name", methodName},
 			},
 		}); err != nil {
		    return
		}
		log.Printf("[goclub_sql migrate]done: " +methodName)
	}
	return
}
