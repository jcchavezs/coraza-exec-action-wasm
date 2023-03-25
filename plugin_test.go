//go:generate tinygo build -o ./testdata/hello-world.wasm -target=wasi ./testdata/hello-world/main.go
package exec

import (
	"testing"

	"github.com/corazawaf/coraza/v3"
	"github.com/stretchr/testify/require"
)

func TestExec(t *testing.T) {
	waf, err := coraza.NewWAF(
		coraza.NewWAFConfig().
			WithDirectives(`
			SecRuleEngine ON
			SecDebugLog /dev/stdout
			SecDebugLogLevel 1
			SecRule RESPONSE_STATUS "@streq 200" "phase:3,exec:./testdata/hello-world.wasm"
		`),
	)
	require.NoError(t, err)

	tx := waf.NewTransaction()
	tx.ProcessRequestHeaders()
	tx.ProcessResponseBody()
	tx.ProcessResponseHeaders(200, "HTTP/1.1")
	tx.ProcessLogging()
	tx.Close()

	t.Fail()
}
