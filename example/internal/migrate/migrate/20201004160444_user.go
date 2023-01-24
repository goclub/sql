package migrate

import (
	"context"
	sq "github.com/goclub/sql"
)

type Migrate struct {
	*sq.Database
}

func (dep Migrate) Migrate20201004160444CreateUserTable() (err error) {
	if _, err = dep.Exec(context.TODO(), `
 	CREATE TABLE user (
 	  id char(36) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
 	  name varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
 	  age int(11) NOT NULL DEFAULT '0',
 	  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
 	  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 	  deleted_at timestamp NULL DEFAULT NULL,
 	  PRIMARY KEY (id),
 	  KEY name (name)
 	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`, nil); err != nil {
	    return
	}
	if _, err = dep.Exec(context.TODO(), `
	CREATE TABLE user_address (
	  user_id char(36) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
	  address varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
	  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	  deleted_at timestamp NULL DEFAULT NULL,
	  PRIMARY KEY (user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`, nil); err != nil {
		return
	}
	return
}
