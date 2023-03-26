package exec

import (
	"context"
	"errors"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/corazawaf/coraza/v3/debuglog"
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
	wazero.CompiledModule
	wazero.Runtime
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

	// We compile the module here and instantiate the module on every request
	// as memory needs to be
	cm, err := r.CompileModule(ctx, wasmSrc)
	if err != nil {
		return err
	}

	e.CompiledModule = cm
	e.Runtime = r

	return nil
}

func (e *Exec) Evaluate(r rules.RuleMetadata, tx rules.TransactionState) {
	logger := tx.DebugLogger().With(
		debuglog.Str("action", "exec"),
		debuglog.Int("rule_id", r.ID()),
	)
	ctx := context.Background()

	m, err := e.Runtime.InstantiateModule(
		ctx,
		e.CompiledModule,
		wazero.NewModuleConfig().
			WithStderr(errWriter{logger}).
			WithStdout(outWriter{logger}),
	)
	if err != nil {
		logger.Error().Err(err).Msg("Action failed")
		return
	}

	defer func() {
		_ = m.Close(ctx)
	}()

	results, err := m.ExportedFunction("exec").Call(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Action failed")
		return
	}

	if len(results) != 1 {
		logger.Error().Msg("Action failed with unexpected number of results")
		return
	}

	execOutputSize := uint32(results[0])
	if execOutputSize == 0 {
		logger.Debug().Msg("Action finished with no ouput")
	} else {
		execOutputPtr := uint32(results[0] >> 32)
		execOutput, ok := m.Memory().Read(execOutputPtr, execOutputSize)
		if !ok {
			logger.Error().Msg("Action failed as could not read output")
			return
		}

		// Displaying this in logger for now. We should be able to record
		// the exec output into the transaction itself.
		logger.Trace().Str("exec_ouput", string(execOutput)).Msg("Action finished")
	}
}

func (e *Exec) Type() rules.ActionType {
	return rules.ActionTypeNondisruptive
}
