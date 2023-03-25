package exec

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/corazawaf/coraza/v3/experimental/plugins"
	"github.com/corazawaf/coraza/v3/rules"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
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

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	mod, err := r.Instantiate(ctx, wasmSrc)
	if err != nil {
		return fmt.Errorf("failed to instansiate wasm script: %v", err)
	}

	e.Function = mod.ExportedFunction("run")
	if e.Function == nil {
		return fmt.Errorf("failed to find exported function 'run'")
	}

	return nil
}

func (e *Exec) Evaluate(r rules.RuleMetadata, tx rules.TransactionState) {
	results, err := e.Function.Call(context.Background())
	if err != nil {
		tx.DebugLogger().Error().Err(err).Msg("exec plugin failed")
	}

	tx.DebugLogger().Trace().Msg(fmt.Sprintf("%v", results))
}

func (e *Exec) Type() rules.ActionType {
	return rules.ActionTypeNondisruptive
}
