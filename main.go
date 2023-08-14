package main

import (
	"github.com/aagoldingay/sb-shovel/cmd"
)

var dir, command, connectionString, queueName, pattern /*, tmpl*/ string
var all, isDlq, delay, help, execute bool
var maxWriteCache int
var commandList = map[string]bool{"config": true, "delete": true, "pull": true, "requeue": true, "send": true, "tidy": true}

var version = "v0.6.2"

// func outputCommands() string {
// 	s := ""
// 	// config
// 	s += "config\n\tpersist Service Bus connection strings to a file in the same location as the executable\n\t"
// 	s += "sb-shovel -cmd config update KEY_NAME KEY_VALUE\n\t"
// 	s += "sb-shovel -cmd config list\n\t"
// 	s += "sb-shovel -cmd config remove KEY_NAME"
// 	s += "\n"

// 	// delete
// 	s += "delete\n\tremove messages from queue\n\t"
// 	s += "requires: -conn, -q\n\toptional: -all, -dlq, -delay\n\t"
// 	s += "WARNING: providing '-all' will delete all messages\n\t"
// 	s += "WARNING: execution without '-delay' may cause issues if you are dealing with extremely large queues"
// 	s += "\n"

// 	// pull
// 	s += "pull\n\tperform local file pull from queue\n\t"
// 	s += "requires: -conn, -q\n\toptional: -dlq, -out-lines\n\t" // , -template
// 	s += "output pattern: 'sb-shovel-output/sb_output_<file_number>'\n\t"
// 	// s += "alter file pattern: -template '{{.SystemProperties.SequenceNumber}} - {{.ID}} - {{.Data | printf \"%s\"}}'"
// 	s += "WARNING: local files with the same naming pattern will be overwritten"
// 	s += "\n"

// 	// requeue
// 	s += "requeue\n\treceive then send messages from one queue to another\n\t"
// 	s += "requires: -conn, -q\n\toptional: -dlq, -all\n\t"
// 	s += "WARNING: providing '-all' will delete all messages"
// 	s += "\n"

// 	// send
// 	s += "send\n\tsend JSON messages to a defined queue from a file\n\t"
// 	s += "requires: -conn, -q, -dir\n\toptional: -dlq\n\t"
// 	s += "WARNING: max read size for a file line is 64*4096 characters\n\t"
// 	s += "WARNING: ensure messages are properly formatted before sending"
// 	s += "\n"

// 	// tidy
// 	s += "tidy\n\tselectively delete messages containing a regex pattern\n\t"
// 	s += "requires: -conn, -q, -pattern\n\toptional: -x\n\t"
// 	s += "WARNING: -x (execute) must be provided to delete any matching messages\n\t"
// 	s += "WARNING: Using this command abandons messages that are not matched.\n\t"
// 	s += "NOTE: refer to the approved syntax: https://github.com/google/re2/wiki/Syntax"
// 	// s += "\n"
// 	return s
// }

func main() {
	cmd.Execute()
}
