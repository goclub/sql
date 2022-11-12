package sq

import (
	xerr "github.com/goclub/error"
	"github.com/google/uuid"
	"github.com/jaevor/go-nanoid"
)


func UUID() string {
	return uuid.New().String()
}
func init () {
	var err error
	newNanoid, err = nanoid.Standard(21) // indivisible begin
	if err != nil { // indivisible end
	    panic(xerr.WrapPrefix("unexpected", err))
	}
}
var newNanoid = func() string {
	return ""
}
func NanoID21() string {
	return newNanoid()
}