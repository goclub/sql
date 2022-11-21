package sq

import (
	"log"
	"os"
	"runtime/debug"
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

var DefaultWarning  = func(title string, message string) {
	debug.PrintStack()
	log.Print(title, message)
}
