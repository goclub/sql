package sq

import (
	"errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"log"
	"regexp"
	"strings"
)

type SQLChecker interface {
	Check(checkSQL []string, execSQL string) (matched bool, message string)
	TrackCheckFail(checkSQL []string, execSQL string, message string)
}

type DefaultSQLChecker struct {
	dmp *diffmatchpatch.DiffMatchPatch
}


func (check DefaultSQLChecker) TrackCheckFail(checkSQL []string, execSQL string, message string)  {
	log.Print("goclub/sql:(SQLChecker)\n",
		message+ "\n",
		"checkSQL:\n"+ strings.Join(checkSQL, "\t\n"),
		"exec sql:\n" + execSQL + "\n",
	)
}

func (check DefaultSQLChecker) Check(checkSQL []string, execSQL string) (matched bool, message string){
	if len(checkSQL) == 0 {
		return true, ""
	}
	if check.dmp == nil {
		check.dmp = diffmatchpatch.New()
	}
	for _, s := range checkSQL {
		if s == execSQL {
			return true, ""
		}
	}

	for _, format := range checkSQL {
		different, err := check.different(execSQL, format) ; if err != nil {
			log.Print(err)
			// 无错误处理
		    return false, err.Error()
		}
		if different.match == true {
			return true, ""
		}
	}
	return false, ""
}

type defaultSQLCheckerDifferent struct {
	match bool
	trimmedSQL string
	trimmedFormat string
}
func (check DefaultSQLChecker) different(execSQL string, format string) (different defaultSQLCheckerDifferent, err error) {
	different.match = true
	trimmedFormat := format
	trimmedSQL := execSQL
	reg, err := regexp.Compile(`\(\?(, \?)*?\)`) ; if err != nil {
		return
	}
	trimmedSQL = reg.ReplaceAllString(trimmedSQL,"")
	trimmedFormat = strings.Replace(trimmedFormat, "{#IN#}", "", -1)
	optional, err := check.matchCheckSQLOptional(format) ; if err != nil {
		return different, err
	}
	for _,optionalItem := range optional {
		trimmedFormat = strings.Replace(trimmedFormat, "{#"+optionalItem+"#}", "", 1)
	}
	for _,optionalItem := range optional {
		trimmedSQL = strings.Replace(trimmedSQL, optionalItem, "", 1)
	}
	if trimmedSQL != trimmedFormat {
		different.match = false
		different.trimmedSQL = trimmedSQL
		different.trimmedFormat = trimmedFormat
		return
	}
	return
}
func (check DefaultSQLChecker)  matchCheckSQLOptional(str string) (optional []string, err error) {
	strLen :=len(str)
	type Position struct {
		Start int
		End int
		Done bool
	}
	data := []Position{}

	for index, s := range str {
		switch s {
		case []rune("{")[0]:
			// last rune
			if index == strLen - 1 { continue }
			nextRune := str[index+1]
			if nextRune == []byte("#")[0] {
				// 闭合检查
				if len(data) != 0 {
					last := data[len(data)-1]
					if last.Done == false {
						message := "goclub/sql: SQLCheck missing #}\n" + str + "\n" +
							strings.Repeat(" ", index) + "^"
						return nil, errors.New(message)
					}
				}
				data = append(data, Position{
					Start: index,
				})
			}
		case []rune("#")[0]:
			// last rune
			if index == strLen - 1 { continue }
			nextRune := str[index+1]
			if nextRune == []byte("}")[0] {
				endIndex := index+2
				// 开启检查
				if len(data) == 0 {
					return nil, errors.New("goclub/sq;: SQLCheck missing {#\n" + str + "\n" +
						strings.Repeat(" ", index) + "^")
				}
				last := data[len(data)-1]
				if last.Done == true {
					message := "goclub/sql: SQLCheck missing {#\n" + str + "\n" +
						strings.Repeat(" ", index) + "^"
					return nil, errors.New(message)
				}
				last.End = endIndex
				last.Done = true
				data[len(data)-1] = last
			}
		}
	}
	for _, item := range data {
		if item.Done == false {
			message := "goclub/sql: SQLCheck missing #}\n" + str + "\n" +
				strings.Repeat(" ", len(str)) + "^"
			return nil, errors.New(message)
		}
		optional = append(optional, str[item.Start+2:item.End-2])
	}
	return
}
