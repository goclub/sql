package sq

import (
	"errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"log"
	"regexp"
	"runtime/debug"
	"strings"
)

type SQLChecker interface {
	Check(checkSQL []string, execSQL string) (pass bool, refs string, err error)
	TrackFail (err error, reviews []string, query string, refs string)
}

type DefaultSQLChecker struct {
	dmp *diffmatchpatch.DiffMatchPatch
}
func (check DefaultSQLChecker) getDmp() *diffmatchpatch.DiffMatchPatch {
	if check.dmp == nil {
		check.dmp = diffmatchpatch.New()
	}
	return check.dmp
}
func (check DefaultSQLChecker) Check(reviews []string, query string) (pass bool, refs string, err error){
	if len(reviews) == 0 {
		return true, "", nil
	}
	for _, s := range reviews {
		if s == query {
			return true, "", nil
		}
	}
	for _, format := range reviews {
		matched, ref, err := check.match(query, format) ; if err != nil {
		    return false, refs, err
		}
		refs += ref
		if matched == true {
			return true, "", nil
		}
	}
	return false, refs, nil
}
func (check DefaultSQLChecker) TrackFail(err error, reviews []string, query string, refs string) {
	defer debug.PrintStack()
	if err != nil {
		log.Print(err)
		return
	}
	message := "query:\n" + query +
		       "reviews:\n" + strings.Join(reviews, "\n")+
		       "refs:\n" + refs
	log.Print(message)
}

type defaultSQLCheckerDifferent struct {
	match bool
	trimmedSQL string
	trimmedFormat string
}
func (check DefaultSQLChecker) match(query string, format string) (matched bool, ref string, err error) {
	trimmedFormat := format

	trimmedSQL := query
	// remove  {#IN#} 和 (?, ?)
	reg, err := regexp.Compile(`\(\?(, \?)*?\)`) ; if err != nil {
		return
	}
	trimmedSQL = reg.ReplaceAllString(trimmedSQL,"")
	trimmedFormat = strings.Replace(trimmedFormat, "{#IN#}", "", -1)
	optional, err := check.matchCheckSQLOptional(trimmedFormat) ; if err != nil {
		return
	}
	for _,optionalItem := range optional {
		trimmedFormat = strings.Replace(trimmedFormat, "{#"+optionalItem+"#}", "", 1)
	}
	for _,optionalItem := range optional {
		trimmedSQL = strings.Replace(trimmedSQL, optionalItem, "", 1)
	}
	trimmedFormat = strings.TrimSpace(trimmedFormat)
	trimmedSQL = strings.TrimSpace(trimmedSQL)
	if trimmedSQL == trimmedFormat {
		return true, "", nil
	}
	ref = "\n\"" + trimmedSQL + "\"\n\"" + trimmedFormat +"\""
	return
}
// 匹配 QB{}.Review 中的 {# AND `name` = ?#} 部分并返回
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
				// 检查之前是否出现 {# 但没有 #} 这种错误
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
				// 检查 #} 之前必须存在 {#
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
