package sq

import (
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestDefaultSQLChecker_Check(t *testing.T) {
	check := DefaultSQLChecker{}
	{
		checkSQL := []string{
			"select * from user where id in {#IN#}",
		}
		{
			execSQL := "select * from user where id in (?)"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.Equal(t, message, "")
		}
		{
			execSQL := "select * from user where id in (?, ?)"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.Equal(t, message, "")
		}
		{
			execSQL := "select * from user where id in (?, ?, ?)"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.Equal(t, message, "")
		}
		{
			execSQL := "select * from user where id in (?, ?, ?, ?)"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.Equal(t, message, "")
		}
	}
	{
		checkSQL := []string{
			"select * from user where mobile = ?{# and name = ?#}{# and age = ?#} limit ?",
		}
		{
			execSQL := "select * from user where mobile = ? limit ?"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.Equal(t, message, "")
		}
		{
			execSQL := "select * from user where mobile = ? and name = ? limit ?"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.Equal(t, message, "")
		}
		{
			execSQL := "select * from user where mobile = ? and age = ? limit ?"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.Equal(t, message, "")
		}
		{
			execSQL := "select * from user where mobile = ? and name = ? and age = ? limit ?"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.Equal(t, message, "")
		}
		{
			execSQL := "select * from user"
			matched, message := check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, false)
			assert.Equal(t, message, "")
		}
	}
}

