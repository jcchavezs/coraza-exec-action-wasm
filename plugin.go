package exec

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/corazawaf/coraza/v3/debuglog"
	"github.com/corazawaf/coraza/v3/experimental/plugins"
	"github.com/corazawaf/coraza/v3/experimental/plugins/plugintypes"
	wazeroapi "github.com/tetratelabs/wazero/api"
)

func init() {
	// Register the plugin
	plugins.RegisterAction("exec", newExec)
}

func newExec() plugintypes.Action {
	return &Exec{}
}

type Exec struct {
	wazero.CompiledModule
	wazero.Runtime
}

type loggerKey struct{}

func putLoggerIntoCtx(ctx context.Context, logger debuglog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func getLoggerFromCtx(ctx context.Context) debuglog.Logger {
	return ctx.Value(loggerKey{}).(debuglog.Logger)
}

var emptyBody = make([]byte, 0)

// mustRead is like api.Memory except that it panics if the offset and byteCount are out of range.
func mustRead(mem wazeroapi.Memory, fieldName string, offset, byteCount uint32) []byte {
	if byteCount == 0 {
		return emptyBody
	}
	buf, ok := mem.Read(offset, byteCount)
	if !ok {
		panic(fmt.Errorf("out of memory reading %s", fieldName))
	}
	return buf
}

// mustReadString is a convenience function that casts mustRead
func mustReadString(mem wazeroapi.Memory, fieldName string, offset, byteCount uint32) string {
	if byteCount == 0 {
		return ""
	}
	return string(mustRead(mem, fieldName, offset, byteCount))
}

func log(ctx context.Context, mod wazeroapi.Module, params []uint64) {
	level := debuglog.Level(params[0])
	message := uint32(params[1])
	messageLen := uint32(params[2])

	var msg string
	if messageLen > 0 {
		msg = mustReadString(mod.Memory(), "message", message, messageLen)
	}

	switch level {
	case debuglog.LevelTrace:
		getLoggerFromCtx(ctx).Trace().Msg(msg)
	case debuglog.LevelDebug:
		getLoggerFromCtx(ctx).Debug().Msg(msg)
	case debuglog.LevelInfo:
		getLoggerFromCtx(ctx).Info().Msg(msg)
	case debuglog.LevelWarn:
		getLoggerFromCtx(ctx).Warn().Msg(msg)
	case debuglog.LevelError:
		getLoggerFromCtx(ctx).Error().Msg(msg)
	}
}

const i32 = wazeroapi.ValueTypeI32

func (e *Exec) Init(rm plugintypes.RuleMetadata, opts string) error {
	if len(opts) == 0 {
		return errors.New("no wasm file provided")
	}

	wasmSrc, err := os.ReadFile(opts)
	if err != nil {
		return err
	}

	ctx := context.Background()
	r := wazero.NewRuntime(ctx)

	hm := r.NewHostModuleBuilder("exec").
		NewFunctionBuilder().
		WithGoModuleFunction(wazeroapi.GoModuleFunc(log), []wazeroapi.ValueType{i32, i32, i32}, []wazeroapi.ValueType{}).
		WithParameterNames("level", "message", "message_len").Export("tx_log")

	hm.Instantiate(ctx)

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

func (e *Exec) Evaluate(r plugintypes.RuleMetadata, tx plugintypes.TransactionState) {
	logger := tx.DebugLogger().With(
		debuglog.Str("action", "exec"),
		debuglog.Int("rule_id", r.ID()),
	)
	ctx := context.Background()

	m, err := e.Runtime.InstantiateModule(
		ctx,
		e.CompiledModule,
		wazero.NewModuleConfig().WithStderr(errWriter{logger}),
	)
	if err != nil {
		logger.Error().Err(err).Msg("Action failed")
		return
	}

	defer func() {
		_ = m.Close(ctx)
	}()

	ctx = putLoggerIntoCtx(ctx, logger)
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

		logger.Debug().Str("exec_output", string(execOutput)).Msg("Action finished")
	}
}

func (e *Exec) Type() plugintypes.ActionType {
	return plugintypes.ActionTypeNondisruptive
}
