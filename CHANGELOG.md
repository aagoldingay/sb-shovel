# 0.2

NEW (internal) [#8](https://github.com/aagoldingay/sb-shovel/issues/8)

- Modularised project structure into packages
- Refactored Service Bus dependency to allow for dependency injection

FIXED

- [#4](https://github.com/aagoldingay/sb-shovel/issues/4) `emptyAll` now removes all queue contents

REMOVED

- Templating functionality for `dump` command (to be solved in [#9](https://github.com/aagoldingay/sb-shovel/issues/9))

# 0.1

Initial version, migrated from local development to open source.

NEW

- Message Peek dump. Pull all messages from a defined queue to a local directory. Includes templated output
    - Default: `{{.Data | printf "%s"}}`
    - More info on how to use templates: https://pkg.go.dev/text/template
    - More info on Service Bus Messages https://pkg.go.dev/github.com/Azure/azure-service-bus-go#Message
- Empty queue. Clear or requeue one or all messages on a queue
- Send messages from a json file