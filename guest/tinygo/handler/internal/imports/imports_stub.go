//go:build !tinygo.wasm

package imports

import "github.com/corazawaf/coraza/v3/debuglog"

// txLog is stubbed for compilation outside TinyGo.
func txLog(level debuglog.Level, messagePtr uintptr, messageLen uint32) {}
