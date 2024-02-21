package handler

import (
	"github.com/corazawaf/coraza/v3/debuglog"
	"github.com/jcchavezs/coraza-exec-wasm/guest/tinygo/handler/internal/imports"
)

// TxLog logs a message at the given level.
func TxLog(level debuglog.Level, message string) {
	messagePtr, messageLen := stringToPtr(message)
	imports.TxLog(level, messagePtr, messageLen)
}
