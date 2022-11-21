package sq

import (
	"log"
	"os"
	"runtime/debug"
)

var Log = log.New(os.Stdout, "goclub/sql: ", log.Ldate|log.Ltime)

func cleanPrint(run func(logger *log.Logger)) {
	flags := Log.Flags()
	prefix := Log.Prefix()
	Log.SetPrefix("")
	Log.SetFlags(0)
	run(Log)
	Log.SetPrefix(prefix)
	Log.SetFlags(flags)
}

var Warning = func(title string, message string) {
	debug.PrintStack()
	log.Print(title, message)
}
