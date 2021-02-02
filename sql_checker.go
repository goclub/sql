package sq

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"log"
	"regexp"
	"strings"
)

type SQLChecker interface {
	Check(checkSQL []string, actual string) (matched bool, diff []string)
	Log(diff []string)
}

type DefaultSQLCheck struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

func (check DefaultSQLCheck) Check(checkSQL []string, actual string) (matched bool, diff []string){
	if check.dmp == nil {
		check.dmp = diffmatchpatch.New()
	}
	for _, s := range checkSQL {
		if s == actual {
			return true, nil
		}

		if strings.HasPrefix(s, "/") && strings.HasSuffix(s, "/") {
			reg, err := regexp.Compile(s) ; if err != nil {
				return false, []string{err.Error()}
			}
			if reg.MatchString(actual) {
				return true, nil
			}
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
	return false, diff
}
func (check DefaultSQLCheck) Log(diff []string)  {
	log.Print(strings.Join(diff, "\n"))
}