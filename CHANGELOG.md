# 0.4

ADDED

- `releaseBundle.sh`
    - Bash script to simplify the production and naming of executables

CHANGED

- Simplified command names
    - `dump` -> `pull`
    - `sendFromFile` -> `send`
    - `emptyOne` -> `delete`
    - `emptyAll` -> `delete -all`
- Split `requeue` functionality from `empty`
    - Fixes [#11](https://github.com/aagoldingay/sb-shovel/issues/11) by separating incompatible concurrency
    - Requeue one message: `requeue`
    - Requeue all messages: `requeue -all`
- Delete and Requeue all progressive output now overwrites the previous line

# 0.3

CHANGED

- Improved `emptyAll`'s speed and capability to delete large quantities of messages
    - WARNING: This is an unbound process without adding the `-delay` flag
    - `delay` flag introduces a 250ms sleep period per 50 messages processed (deleted or requeued)

# 0.2.2

NEW

- Documented `go install` option, now that packages are imported appropriately

FIXED

- `emptyAll` now correctly prints status messages per 50 deleted files, and does not exit the process

CHANGED

- Moved version to a variable to make it slightly easier to remember

# 0.2.1

CHANGED

- Referencing own GitHub repo hostname when importing

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