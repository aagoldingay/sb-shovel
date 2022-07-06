package main

import (
	"flag"
	"fmt"

	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
)

var dir, command, connectionString, queueName /*, tmpl*/ string
var all, isDlq, delay, help bool
var maxWriteCache int
var commandList = map[string]bool{"empty": true, "pull": true, "requeue": true, "send": true}

var version = "v0.4"

func outputCommands() string {
	s := ""
	// empty
	s += "emptyAll\n\tremove messages from queue\n\t"
	s += "requires: -conn, -q\n\toptional: -all, -dlq, -delay\n\t"
	s += "WARNING: providing '-all' will delete all messages\n\t"
	s += "WARNING: execution without '-delay' may cause issues if you are dealing with extremely large queues"
	s += "\n"

	// pull
	s += "pull\n\tperform local file pull from queue\n\t"
	s += "requires: -conn, -q\n\toptional: -dlq, -out-lines\n\t" // , -template
	s += "output pattern: 'sb-shovel-output/sb_output_<file_number>'\n\t"
	// s += "alter file pattern: -template '{{.SystemProperties.SequenceNumber}} - {{.ID}} - {{.Data | printf \"%s\"}}'"
	s += "WARNING: local files with the same naming pattern will be overwritten"
	s += "\n"

	// requeue
	s += "requeue\n\treceive then send messages from one queue to another\n\t"
	s += "requires: -conn, -q\n\toptional: -dlq, -all\n\t"
	s += "WARNING: providing '-all' will delete all messages"
	s += "\n"

	// send
	s += "send\n\tsend JSON messages to a defined queue from a file\n\t"
	s += "requires: -conn, -q, -dir\n\toptional: -dlq\n\t"
	s += "WARNING: max read size for a file line is 64*4096 characters\n\t"
	s += "WARNING: ensure messages are properly formatted before sending"
	//s += "\n"
	return s
}

func main() {
	flag.StringVar(&connectionString, "conn", "", "service bus connection string\ne.g. \"Endpoint=sb://<service_bus>.servicebus.windows.net/;SharedAccessKeyName=<key_name>;SharedAccessKey=<key_value>\"")
	flag.StringVar(&queueName, "q", "", "service bus queue name")
	flag.StringVar(&command, "cmd", "", outputCommands())
	// flag.StringVar(&tmpl, "template", `{{.Data | printf "%s"}}`, "template syntax: https://pkg.go.dev/text/template\nmessage attributes: see https://pkg.go.dev/github.com/Azure/azure-service-bus-go#Message")
	flag.StringVar(&dir, "dir", "", "directory of file containing json messages to send")
	flag.BoolVar(&all, "all", false, "perform the operation on an entire entity")
	flag.BoolVar(&isDlq, "dlq", false, "point to the defined queue's deadletter subqueue")
	// flag.BoolVar(&requeue, "rq", false, "resubmit (requeue) messages")
	flag.BoolVar(&delay, "delay", false, "include a 250ms delay for every 50 messages sent")
	flag.BoolVar(&help, "help", false, "information about this tool")
	flag.IntVar(&maxWriteCache, "out-lines", 100, "number of lines per file")
	flag.Parse()

	if _, cmdPres := commandList[command]; !cmdPres || help || (cmdPres && (len(connectionString) == 0 || len(queueName) == 0)) {
		fmt.Printf("sb-shovel %s\nManage large message operations on a given Service Bus.\n\n", version)
		fmt.Println("Example Usage:\n\tsb-shovel.exe -cmd dump -conn \"<servicebus_uri>\" -q queueName\n\tsb-shovel.exe -cmd emptyAll -conn \"<servicebus_uri>\" -q queueName -dlq -rq")
		flag.PrintDefaults()
		return
	}

	sb, err := sbc.NewController(connectionString)
	if err != nil {
		fmt.Println(err)
	}

	switch command {
	case "pull":
		if maxWriteCache < 1 {
			fmt.Println("Value for -out-lines is not valid. Must be >= 1")
			return
		}
		if delay {
			fmt.Println("Delay is not supported for this command")
			return
		}
		err := pull(sb, queueName, isDlq, maxWriteCache)
		if err != nil {
			fmt.Println(err)
		}
		return
	case "empty":
		if delay {
			fmt.Println("Delay is not supported for this command")
			return
		}
		err := empty(sb, queueName, isDlq, all, delay)
		// err := empty(sb, queueName, isDlq, false, requeue, false)
		if err != nil {
			fmt.Println(err)
		}
		return
	case "requeue":
		// err := empty(sb, queueName, isDlq, true, requeue, delay)
		if delay {
			fmt.Println("Delay is not supported for this command")
			return
		}
		err := requeue(sb, queueName, all, isDlq)
		if err != nil {
			fmt.Println(err)
		}
		return
	case "send":
		if len(dir) == 0 {
			fmt.Println("Value for -dir flag missing")
			return
		}
		if isDlq {
			fmt.Println("Cannot send to a dead letter queue")
			return
		}
		if delay {
			fmt.Println("Delay is not supported for this command")
			return
		}
		err := sendFromFile(sb, queueName, dir)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}
