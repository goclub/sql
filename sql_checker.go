package sq

import (
	xerr "github.com/goclub/error"
	"github.com/sergi/go-diff/diffmatchpatch"
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
		DefaultLog.Print(err)
		return
	}
	message := "query:\n" + query + "\n" +
		       "reviews:\n" + strings.Join(reviews, "\n")+ "\n" +
		       "refs:\n" + refs
	DefaultLog.Print(message)
}

type defaultSQLCheckerDifferent struct {
	match bool
	trimmedSQL string
	trimmedFormat string
}
func (check DefaultSQLChecker) match(query string, format string) (matched bool, ref string, err error) {
	trimmedFormat := format

	trimmedSQL := query
	// remove {#VALUES#} 和 (?,?),(?,?) 和 (?,?)
	{
		var reg *regexp.Regexp
		reg, err = regexp.Compile(`VALUES \(.*\)`) ; if err != nil {
		return
	}
		trimmedSQL = reg.ReplaceAllString(trimmedSQL,"VALUES ")
		trimmedFormat = strings.Replace(trimmedFormat, "{#VALUES#}", "", -1)
	}
	// remove  {#IN#} 和 (?,?)
	{
		var reg *regexp.Regexp
		reg, err = regexp.Compile(`\(\?(,\?)*?\)`) ; if err != nil {
		return
	}
		trimmedSQL = reg.ReplaceAllString(trimmedSQL,"")
		trimmedFormat = strings.Replace(trimmedFormat, "{#IN#}", "", -1)
	}
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
	ref = "\n   sql: \"" + trimmedSQL + "\"\nformat: \"" + trimmedFormat +"\""
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
						return nil, xerr.New(message)
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
					return nil, xerr.New("goclub/sq;: SQLCheck missing {#\n" + str + "\n" +
						strings.Repeat(" ", index) + "^")
				}
				last := data[len(data)-1]
				if last.Done == true {
					message := "goclub/sql: SQLCheck missing {#\n" + str + "\n" +
						strings.Repeat(" ", index) + "^"
					return nil, xerr.New(message)
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
			return nil, xerr.New(message)
		}
		optional = append(optional, str[item.Start+2:item.End-2])
	}
	return
}
