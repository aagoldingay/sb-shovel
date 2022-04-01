# sb-shovel

sb-shovel is a CLI tool, written in Go, designed to speed up processing Azure Service Bus queue messages. 

This began as a locally developed tool to assist in performing operations on dead-letter queues, in Production environments, with large message volumes. It has proven to be invaluable for myself, so here's hoping somebody else will find this useful!

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

## Contributing

I welcome anybody to contribute if they have something to add or improve! 

If you spot a problem while using this tool, please open an issue - that alone would be a great help!

More in [contributing.md](/CONTRIBUTING.md) and [code of conduct](/CODE_OF_CONDUCT.md).

## License
[MIT](https://choosealicense.com/licenses/mit/)