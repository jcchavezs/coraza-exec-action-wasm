package imports

import "github.com/corazawaf/coraza/v3/debuglog"

func TxLog(level debuglog.Level, messagePtr uintptr, messageLen uint32) {
	txLog(level, messagePtr, messageLen)
}
