package main

import (
	"unsafe"

	"github.com/corazawaf/coraza/v3/debuglog"
	"github.com/jcchavezs/coraza-exec-wasm/guest/tinygo/handler"
)

//export exec
func exec() uint64 {
	handler.TxLog(debuglog.LevelDebug, "This log shows up as DEBUG level!")

	outputPtr, outputSize := stringToPtr("Hello world!")
	// TODO(jcchavezs): Should I export this function in the main package
	// so userland code does not do copy-pasta?
	return (uint64(outputPtr) << 32) | uint64(outputSize)
}

// main is required for the `wasi` target, even if it isn't used.
// See https://wazero.io/languages/tinygo/#why-do-i-have-to-define-main
func main() {}

// stringToPtr returns a pointer and size pair for the given string
// in a way that is compatible with WebAssembly numeric types.
func stringToPtr(s string) (uint32, uint32) {
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(buf))
}
