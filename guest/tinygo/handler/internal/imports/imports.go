//go:build tinygo.wasm

package imports

import "github.com/corazawaf/coraza/v3/debuglog"

//go:wasm-module exec
//go:export tx_log
func txLog(level debuglog.Level, messagePtr uintptr, messageLen uint32)
