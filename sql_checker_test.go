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
			matched, _,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
		{
			execSQL := "select * from user where id in (?, ?)"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
		{
			execSQL := "select * from user where id in (?, ?, ?)"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
		{
			execSQL := "select * from user where id in (?, ?, ?, ?)"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
	}
	{
		checkSQL := []string{
			"select * from user where id in {#IN#} limit ?",
		}
		{
			execSQL := "select * from user where id in (?) limit ?"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
		{
			execSQL := "select * from user where id in (?, ?) limit ?"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
	}
	{
		checkSQL := []string{
			"select * from user where mobile = ?{# and name = ?#}{# and age = ?#} limit ?",
		}
		{
			execSQL := "select * from user where mobile = ? limit ?"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
		{
			execSQL := "select * from user where mobile = ? and name = ? limit ?"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
		{
			execSQL := "select * from user where mobile = ? and age = ? limit ?"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
		{
			execSQL := "select * from user where mobile = ? and name = ? and age = ? limit ?"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
		{
			execSQL := "select * from user"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, false)
			assert.NoError(t, err)
		}
	}
}

func TestDefaultSQLChecker_Check2(t *testing.T) {
	check := DefaultSQLChecker{}
	{
		checkSQL := []string{
			"select * from user where mobile = ? {#and name = ?#} limit ?",
		}
		{
			execSQL := "select * from user where mobile = ? limit ?"
			matched,_,  err:= check.Check(checkSQL, execSQL)
			assert.Equal(t, matched, true)
			assert.NoError(t, err)
		}
	}
}
