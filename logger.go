package sq

import (
	"log"
	"os"
)

var DefaultLog = log.New(os.Stdout, "", log.Ldate|log.Ltime)

