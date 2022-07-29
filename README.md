# sb-shovel

sb-shovel is a CLI tool, written in Go, designed to speed up processing Azure Service Bus queue messages. 

This began as a locally developed tool to assist in performing operations on dead-letter queues, in Production environments, with large message volumes. It has proven to be invaluable for myself, so here's hoping somebody else will find this useful!

## Usage

Run one of the following commands to get a complete breakdown of the available commands:

```
sb-shovel.exe
sb-shovel.exe -help
```

Peek-dump queue contents to a local directory:

```
sb-shovel.exe -cmd pull -conn "<servicebus_connection_string>" -q queueName
```

Purge the contents of an entire dead-letter queue

```
sb-shovel.exe -cmd delete -conn "<servicebus_connection_string>" -q queueName -dlq -all
```

## Installation and Running

Install and set up your Go (1.17+) environment (see main README)

Once done, you can now clone to GOROOT or [GOPATH](https://www.digitalocean.com/community/tutorials/understanding-the-gopath), and set up the environment

Easy: 
```
go install github.com/aagoldingay/sb-shovel
```

Git clone:
```
cd $GOPATH/src
git clone <repo>
cd sb-shovel
go mod init
go mod download
go run main.go commands.go
```

## Project Structure

```
sb-shovel
│   .gitignore
│   CHANGELOG.md
│   CODE_OF_CONDUCT.md
│   commands.go
│   commands_test.go
│   CONTRIBUTING.md
│   go.mod
│   go.sum
│   LICENSE
│   main.go
│   README.md
|   releaseBundle.sh
│
├───.github                             # repository configurations
│   ├───ISSUE_TEMPLATE
│   │       bug_report.md
│   │       feature_request.md
│   │
│   └───workflows
│           go_ci.yml
│
├───config
│       config.go
|
├───io
│       files.go
│       files_test.go
│
├───mocks
│       mockcontroller.go
│       mockcontroller_test.go
│
├───sbcontroller
│       controller.go
│       controller_integration_test.go
│
├───test_files                          # files to support project testing
│       cmd_send_test.txt
│       filewriter_test.json
│       integration_template.json
│
└───test_terraform                      # terraform configuration for project testing
    │   providers.tf
    │   README.md
    │   servicebus.tf
    │   variables.tf
```

## Building
In the same directory as main.go, this is all you need:

```
go build
```

## Testing

There are some fairly fundamental unit tests in this project, which can be run with `go test`. 

To run integration tests on sbcontroller module, you will need to configure `test_file/integration.json` using `test_files/integration_template.json` file as a guide. Those tests designed to interact directly with a real Azure Service Bus resource will not run on the CI pipeline.

The `test_terraform` directory, and it's accompanying [README](test_terraform/README.md), offer a basic configuration for an Azure Service Bus resource to support user testing. 

## Contributing

I welcome anybody to contribute if they have something to add or improve! 

If you spot a problem while using this tool, please open an issue - that alone would be a great help!

More in [contributing.md](/CONTRIBUTING.md) and [code of conduct](/CODE_OF_CONDUCT.md).

## License
[MIT](https://choosealicense.com/licenses/mit/)