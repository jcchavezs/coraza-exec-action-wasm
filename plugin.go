package exec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/corazawaf/coraza/v3/experimental/plugins"
	"github.com/corazawaf/coraza/v3/rules"
)

func init() {
	// Register the plugin
	plugins.RegisterAction("exec", newExec)
}

func newExec() rules.Action {
	return &Exec{}
}

type Exec struct {
	api.Function
	api.Memory
}

func (e *Exec) Init(rm rules.RuleMetadata, opts string) error {
	if len(opts) == 0 {
		return errors.New("no wasm file provided")
	}

	wasmSrc, err := os.ReadFile(opts)
	if err != nil {
		return err
	}
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return err
	}

	// TODO(jcchavezs): Should I reuse this module given that the request is keeping
	// some values in memory to send them back from guest to host?
	mod, err := r.Instantiate(ctx, wasmSrc)
	if err != nil {
		return fmt.Errorf("failed to instansiate wasm script: %v", err)
	}

	e.Function = mod.ExportedFunction("run")
	if e.Function == nil {
		return fmt.Errorf("failed to find exported function 'run'")
	}
	// TODO(jcchavezs):
	// - should I make the memory stateful?
	// - how do I clean the memory after the request is gone?
	e.Memory = mod.Memory()

	return nil
}

func (e *Exec) Evaluate(_ rules.RuleMetadata, tx rules.TransactionState) {
	results, err := e.Function.Call(context.Background())
	if err != nil {
		tx.DebugLogger().Error().Err(err).Msg("exec plugin failed")
	}

	if len(results) != 1 {
		tx.DebugLogger().Error().Msg("exec plugin failed, no results")
		return
	}

	execOutputPtr := uint32(results[0] >> 32)
	execOutputSize := uint32(results[0])
	execOutput, ok := e.Memory.Read(execOutputPtr, execOutputSize)
	if !ok {
		tx.DebugLogger().Error().Msg("exec plugin failed, could not read output")
		return
	}
	tx.DebugLogger().Trace().Msg(string(execOutput))
}

func (e *Exec) Type() rules.ActionType {
	return rules.ActionTypeNondisruptive
}

func ptrToString(ptr uint32, size uint32) (ret string) {
	// Here, we want to get a string represented by the ptr and size. If we
	// wanted a []byte, we'd use reflect.SliceHeader instead.
	strHdr := (*reflect.StringHeader)(unsafe.Pointer(&ret))
	strHdr.Data = uintptr(ptr)
	strHdr.Len = int(size)
	return
}
