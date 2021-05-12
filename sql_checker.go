package sq

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"log"
	"runtime/debug"
	"strings"
)

type SQLChecker interface {
	Check(checkSQL []string, actual string) (matched bool, diff []string, stack []byte)
	TrackCheckFail(diff []string, stack []byte)
}

type DefaultSQLChecker struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

func (check DefaultSQLChecker) Check(checkSQL []string, actual string) (matched bool, diff []string, stack []byte){
	if len(checkSQL) == 0 {
		return true, nil, nil
	}
	if check.dmp == nil {
		check.dmp = diffmatchpatch.New()
	}
	for _, s := range checkSQL {
		if s == actual {
			return true, nil, nil
		}
	}
	diff = []string{
		"expected:" + strings.Join(checkSQL, " "),
		"actual:"+actual,
	}

	for _, s := range checkSQL {
		result := check.dmp.DiffMain(s, actual, false)
		diff = append(diff, check.dmp.DiffPrettyText(result))
	}
	return false, diff, debug.Stack()
}

func (check DefaultSQLChecker) TrackCheckFail(diff []string, stack []byte)  {
	log.Print("goclub/sql:(SQLChecker)\n",
		strings.Join(diff, "\n"), "\n",
		string(stack), )
}
