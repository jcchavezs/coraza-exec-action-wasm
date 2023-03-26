# Coraza `exec` action WASM

`exec` action plugin for coraza running over WASM. It is inspired on the [exec action in modsecurity](https://github.com/SpiderLabs/ModSecurity/wiki/Reference-Manual-%28v3.x%29#exec) but targets safer execution.

Main difference between `modsecurity` and the `coraza` exec action is that this plugin runs in a safe environment over [wazero](https://wazero.io) runtime and can't perform actions in the host and instead targets to record the execution output in the transaction itself as a variable.

## Getting started

When in a ruleset like:

```seclang
SecRuleEngine ON
SecDebugLog /dev/stdout
SecDebugLogLevel 1
SecRule RESPONSE_STATUS "@streq 200" "phase:3,exec:./testdata/hello-world.wasm"
```

The debug log will show:

```console
2023/03/26 18:28:59 [DEBUG] Matching rule tx_id="qpVlKXOQCXRluSAMnSP" rule_id=123 variable_name="RESPONSE_STATUS" key=""
2023/03/26 18:28:59 [DEBUG] Evaluating action tx_id="qpVlKXOQCXRluSAMnSP" action="exec"
2023/03/26 18:28:59 [DEBUG] This log shows up as DEBUG level! tx_id="qpVlKXOQCXRluSAMnSP" action="exec" rule_id=123
2023/03/26 18:28:59 [TRACE] Action finished tx_id="qpVlKXOQCXRluSAMnSP" action="exec" rule_id=123 exec_ouput="Hello world!"
```

## TODO

- Be able to load the WASM file from the WAF RootFS.
- Record the exec output into `EXEC_OUTPUT` variable in the transaction.
- Decide what parameters are passed to the exec call as env vars (double check with modsecurity).
