package sq

import (
	"log"
	"os"
)

var DefaultLog = log.New(os.Stdout, "goclub/sql: ", log.Ldate|log.Ltime)

func cleanPrint(run func(logger *log.Logger)) {
	flags := DefaultLog.Flags()
	prefix := DefaultLog.Prefix()
	DefaultLog.SetPrefix("")
	DefaultLog.SetFlags(0)
	run(DefaultLog)
	DefaultLog.SetPrefix(prefix)
	DefaultLog.SetFlags(flags)
}
