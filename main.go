package main

import (
	"flag"
	"fmt"

	sbc "sb-shovel/sbcontroller"
)

var dir, command, connectionString, queueName /*, tmpl*/ string
var isDlq, requeue, help bool
var maxWriteCache int
var commandList = map[string]string{"dump": "dump", "emptyOne": "emptyOne", "emptyAll": "emptyAll", "sendFromFile": "sendFromFile"}

func outputCommands() string {
	s := ""
	// dump
	s += "dump\n\tperform local file dump from queue\n\t"
	s += "requires: -conn, -q\n\toptional: -dlq, -out-lines\n\t" // , -template
	s += "output pattern: 'sb-shovel-output/sb_output_<file_number>'\n\t"
	// s += "alter file pattern: -template '{{.SystemProperties.SequenceNumber}} - {{.ID}} - {{.Data | printf \"%s\"}}'"
	s += "WARNING: local files with the same naming pattern will be overwritten"
	s += "\n"

	//emptyOne
	s += "emptyOne\n\treceive and remove one message from queue\n\t"
	s += "requires: -conn, -q\n\toptional: -dlq, -rq\n\t"
	s += "WARNING: failure to provide '-rq' will result in lost messages\n\t"
	s += "WARNING: providing -rq while actioning the main queue will point message resubmission to the dead letter queue"
	s += "\n"

	//emptyAll
	s += "emptyAll\n\treceive and remove all messages from queue\n\t"
	s += "requires: -conn, -q\n\toptional: -dlq, -rq\n\t"
	s += "WARNING: failure to provide '-rq' will result in lost messages\n\t"
	s += "WARNING: providing -rq while actioning the main queue will point message resubmission to the dead letter queue"
	s += "\n"

	//sendFromFile
	s += "sendFromFile\n\tsend JSON messages to a defined queue from a file\n\t"
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
	flag.BoolVar(&isDlq, "dlq", false, "point to the defined queue's deadletter subqueue")
	flag.BoolVar(&requeue, "rq", false, "resubmit (requeue) messages")
	flag.BoolVar(&help, "help", false, "information about this tool")
	flag.IntVar(&maxWriteCache, "out-lines", 100, "number of lines per file")
	flag.Parse()

	if _, cmdPres := commandList[command]; !cmdPres || help || (cmdPres && (len(connectionString) == 0 || len(queueName) == 0)) {
		fmt.Println("sb-shovel v0.2\nManage large message operations on a given Service Bus.\n ")
		fmt.Println("Example Usage:\n\tsb-shovel.exe -cmd dump -conn \"<servicebus_uri>\" -q queueName\n\tsb-shovel.exe -cmd emptyAll -conn \"<servicebus_uri>\" -q queueName -dlq -rq")
		flag.PrintDefaults()
		return
	}

	sb, err := sbc.NewController(connectionString)
	if err != nil {
		fmt.Println(err)
	}

	switch command {
	case commandList["dump"]:
		if maxWriteCache < 1 {
			fmt.Println("Value for -out-lines is not valid. Must be >= 1")
			return
		}
		err := dump(sb, queueName, isDlq, maxWriteCache)
		if err != nil {
			fmt.Println(err)
		}
		return
	case commandList["emptyOne"]:
		err := empty(sb, queueName, isDlq, false, requeue)
		if err != nil {
			fmt.Println(err)
		}
		return
	case commandList["emptyAll"]:
		err := empty(sb, queueName, isDlq, true, requeue)
		if err != nil {
			fmt.Println(err)
		}
		return
	case commandList["sendFromFile"]:
		if len(dir) == 0 {
			fmt.Println("Value for -dir flag missing")
			return
		}
		if isDlq {
			fmt.Println("Cannot send to a dead letter queue")
			return
		}
		err := sendFromFile(sb, queueName, dir)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}
