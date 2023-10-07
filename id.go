package sq

import (
	xerr "github.com/goclub/error"
	"github.com/google/uuid"
	"github.com/jaevor/go-nanoid"
	"strings"
)

func UUID() string {
	return uuid.New().String()
}
func UUID32() string {
	return strings.ReplaceAll(UUID(), "-", "")
}
func init() {
	var err error
	newNanoid, err = nanoid.Custom("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789", 24) // indivisible begin
	if err != nil {                                                                                      // indivisible end
		panic(xerr.WrapPrefix("unexpected", err))
	}
}

var newNanoid = func() string {
	panic("unexpected")
}

// NanoID24  `A-Za-z0-9` 24
// 某些第三方接口需要外部订单号是大小写字母加数字,所以用`A-Za-z0-9` 24 比 默认的21更稳妥.
func NanoID24() string {
	return newNanoid()
}
