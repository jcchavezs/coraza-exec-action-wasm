package main

import "unsafe"

//export run
func run() uint64 {
	ptr, size := stringToPtr("Hello world!")
	return (uint64(ptr) << 32) | uint64(size)
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
