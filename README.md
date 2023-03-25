# Coraza exec action WASM

`exec` action plugin for coraza running over WASM. It is somehow inspired on the [exec action in modsecurity](https://github.com/SpiderLabs/ModSecurity/wiki/Reference-Manual-%28v3.x%29#exec).

One difference between modsecurity and the coraza exec action is that this plugin runs in a safe environment over wazero runtime and can't perform actions in the host so it aims to record the output in a transaction variable (still WIP).

## Getting started

```modsecurity
SecRuleEngine ON
SecDebugLog /dev/stdout
SecDebugLogLevel 1
SecRule RESPONSE_STATUS "@streq 200" "phase:3,exec:./testdata/hello-world.wasm"
```
