# 0.1

Initial version, migrated from local development to open source.

NEW

- Message Peek dump. Pull all messages from a defined queue to a local directory. Includes templated output
    - Default: `{{.Data | printf "%s"}}`
    - More info on how to use templates: https://pkg.go.dev/text/template
    - More info on Service Bus Messages https://pkg.go.dev/github.com/Azure/azure-service-bus-go#Message
- Empty queue. Clear or requeue one or all messages on a queue
- Send messages from a json file