package main

import (
	"flag"
	"fmt"
	"strings"

	cc "github.com/aagoldingay/sb-shovel/config"
	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
)

var dir, command, connectionString, queueName, pattern /*, tmpl*/ string
var all, isDlq, delay, help, execute bool
var maxWriteCache int
var commandList = map[string]bool{"config": true, "delete": true, "pull": true, "requeue": true, "send": true, "tidy": true}

var version = "unreleased"

func outputCommands() string {
	s := ""
	// config
	s += "config\n\tpersist Service Bus connection strings to a file in the same location as the executable\n\t"
	s += "sb-shovel -cmd config update KEY_NAME KEY_VALUE\n\t"
	s += "sb-shovel -cmd config list\n\t"
	s += "sb-shovel -cmd config remove KEY_NAME"
	s += "\n"

	// delete
	s += "delete\n\tremove messages from queue\n\t"
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
	s += "\n"

	// tidy
	s += "tidy\n\tselectively delete messages containing a regex pattern\n\t"
	s += "requires: -conn, -q, -pattern\n\toptional: -x\n\t"
	s += "WARNING: -x (execute) must be provided to delete any matching messages\n\t"
	s += "WARNING: Using this command abandons messages that do not match, or matched messages when `-x` is not provided\n\t"
	s += "NOTE: refer to the approved syntax: https://github.com/google/re2/wiki/Syntax"
	// s += "\n"
	return s
}

func checkIfConfig(s string) (bool, string) {
	if strings.HasPrefix(s, "cfg|") {
		return true, strings.Split(s, "|")[1]
	}
	return false, ""
}

func main() {
	flag.StringVar(&connectionString, "conn", "", "service bus connection string\ne.g. \"Endpoint=sb://<service_bus>.servicebus.windows.net/;SharedAccessKeyName=<key_name>;SharedAccessKey=<key_value>\"")
	flag.StringVar(&queueName, "q", "", "service bus queue name")
	flag.StringVar(&command, "cmd", "", outputCommands())
	flag.StringVar(&pattern, "pattern", "", "regex pattern to match against message contents")
	// flag.StringVar(&tmpl, "template", `{{.Data | printf "%s"}}`, "template syntax: https://pkg.go.dev/text/template\nmessage attributes: see https://pkg.go.dev/github.com/Azure/azure-service-bus-go#Message")
	flag.StringVar(&dir, "dir", "", "directory of file containing json messages to send")
	flag.BoolVar(&all, "all", false, "perform the operation on an entire entity")
	flag.BoolVar(&isDlq, "dlq", false, "point to the defined queue's deadletter subqueue")
	flag.BoolVar(&execute, "x", false, "tidy command: perform delete operation")
	flag.BoolVar(&delay, "delay", false, "include a 250ms delay for every 50 messages sent")
	flag.BoolVar(&help, "help", false, "information about this tool")
	flag.IntVar(&maxWriteCache, "out-lines", 100, "number of lines per file")
	flag.Parse()
	args := flag.Args()

	if _, cmdPres := commandList[command]; !cmdPres || help ||
		(cmdPres && command != "config" && (len(connectionString) == 0 || len(queueName) == 0)) && len(args) > 0 ||
		(cmdPres && command == "config" && len(args) == 0) {
		fmt.Printf("sb-shovel %s\nManage large message operations on a given Service Bus.\n\n", version)
		fmt.Println("Example Usage:\n\tsb-shovel.exe -cmd pull -conn \"<servicebus_connectionstring>\" -q queueName\n\tsb-shovel.exe -cmd delete -conn \"<servicebus_connectionstring>\" -q queueName -dlq")
		flag.PrintDefaults()
		return
	}

	cfg, err := cc.NewConfigController("sb-shovel")
	if err != nil {
		fmt.Println(err)
	}

	var sb sbc.Controller

	if command != "config" {
		if isConfig, key := checkIfConfig(connectionString); isConfig {
			err := cfg.LoadConfig()
			if err != nil && err.Error() != cc.ERR_NOCONFIG {
				fmt.Println(err)
				return
			}
			connectionString, err = cfg.GetConfigValue(key)
			if err != nil || connectionString == "" {
				fmt.Println("attribute not found in config.")
				return
			}
			fmt.Printf("connecting to %s\n", key)
		}
		sb, err = sbc.NewServiceBusController(connectionString)
		if err != nil {
			fmt.Println(err)
		}
	}

	switch command {
	case "config":
		if delay {
			fmt.Println("-delay is not supported for this command")
			return
		}
		if isDlq {
			fmt.Println("-dlq is not supported by this command")
			return
		}
		if dir != "" {
			fmt.Println("-dir is not supported by this command")
			return
		}
		config(cfg, args)
		return
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
	case "delete":
		if delay {
			fmt.Println("Delay is not supported for this command")
			return
		}
		err := delete(sb, queueName, isDlq, all, delay)
		if err != nil {
			fmt.Println(err)
		}
		return
	case "requeue":
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
	case "tidy":
		if dir != "" {
			fmt.Println("-dir is not supported by this command")
			return
		}
		if maxWriteCache < 1 {
			fmt.Println("Value for -out-lines is not valid. Must be >= 1")
			return
		}
		if delay {
			fmt.Println("Delay is not supported for this command")
			return
		}
		if len(pattern) == 0 {
			fmt.Println("Pattern must be specified, else all messages risk being deleted")
			return
		}
		err := tidy(sb, queueName, pattern, isDlq, execute)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("finished processing messages")
		return
	}
}
