package sq

import (
	"log"
	"os"
)

var DefaultLog = log.New(os.Stdout, "goclub/sql:", log.Ldate|log.Ltime|log.Lshortfile)

