# sb-shovel

sb-shovel is a CLI tool, written in Go, designed to speed up processing Azure Service Bus queue messages. 

This began as a locally developed tool to assist in performing operations on dead-letter queues, in Production environments, with large message volumes. It has proven to be invaluable for myself, so here's hoping somebody else will find this useful!

## Project Structure

```
sb-shovel
│   .gitignore
│   CHANGELOG.md
│   CODE_OF_CONDUCT.md
│   CONTRIBUTING.md
│   filewriter.go
│   filewriter_test.go
│   go.mod
│   go.sum
│   LICENSE
│   main.go
│   README.md
│   servicebus.go
│
├───.github             # repository configurations
│   ├───ISSUE_TEMPLATE
│   └───workflows
│
├───test_files          # files to support project testing
│
└───test_terraform      # terraform configuration for project testing
    │   providers.tf
    │   README.md
    │   servicebus.tf
    │   variables.tf
```

## Installation and Running

Install and set up your Go (1.17+) environment (see main README)

Once done, you can now clone to GOROOT or [GOPATH](https://www.digitalocean.com/community/tutorials/understanding-the-gopath), and set up the environment

```
cd $GOPATH/src
git clone <repo>
cd sb-shovel
go mod init
go mod download
go run main.go filewriter.go servicebus.go
```

## Building
In the same directory as main.go, this is all you need:

```
go build
```

## Usage

Run one of the following commands to get a complete breakdown of the available commands:

```
sb-shovel.exe
sb-shovel.exe -help
```

Peek-dump queue contents to a local directory:

```
sb-shovel.exe -cmd dump -conn "<servicebus_uri>" -q queueName
```

Purge the contents of an entire dead-letter queue

```
sb-shovel.exe -cmd emptyAll -conn "<servicebus_uri>" -q queueName -dlq
```

## Testing

There are some fairly fundamental unit tests in this project, which can be run with `go test`. 

Unfortunately, the package used to interact with Azure Service Bus does not offer interfaces for Dependency Injection, and thus more work is required to refactor existing logic to support automated testing of the wider tool, without interacting with an Azure Service Bus resource.

The `test_terraform` directory, and it's accompanying [README](test_terraform/README.md), offer a basic configuration for an Azure Service Bus resource to support user testing. 

## Contributing

I welcome anybody to contribute if they have something to add or improve! 

If you spot a problem while using this tool, please open an issue - that alone would be a great help!

More in [contributing.md](/CONTRIBUTING.md) and [code of conduct](/CODE_OF_CONDUCT.md).

## License
[MIT](https://choosealicense.com/licenses/mit/)