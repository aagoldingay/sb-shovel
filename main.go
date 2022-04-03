package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"text/template"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
)

var dir, command, connectionString, queueName, tmpl string
var isDlq, requeue, help, isEmpty bool
var maxWriteCache int
var commandList = map[string]string{"dump": "dump", "emptyOne": "emptyOne", "emptyAll": "emptyAll", "sendFromFile": "sendFromFile"}

func outputCommands() string {
	s := ""
	// dump
	s += "dump\n\tperform local file dump from queue\n\t"
	s += "requires: -conn, -q\n\toptional: -dlq, -out-lines, -template\n\t"
	s += "output pattern: 'sb-shovel-output/sb_output_<file_number>'\n\t"
	s += "alter file pattern: -template '{{.SystemProperties.SequenceNumber}} - {{.ID}} - {{.Data | printf \"%s\"}}'"
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
	flag.StringVar(&tmpl, "template", `{{.Data | printf "%s"}}`, "template syntax: https://pkg.go.dev/text/template\nmessage attributes: see https://pkg.go.dev/github.com/Azure/azure-service-bus-go#Message")
	flag.StringVar(&dir, "dir", "", "directory of file containing json messages to send")
	flag.BoolVar(&isDlq, "dlq", false, "point to the defined queue's deadletter subqueue")
	flag.BoolVar(&requeue, "rq", false, "resubmit (requeue) messages")
	flag.BoolVar(&help, "help", false, "information about this tool")
	flag.IntVar(&maxWriteCache, "out-lines", 100, "number of lines per file")
	flag.Parse()

	if _, cmdPres := commandList[command]; !cmdPres || help || (cmdPres && (len(connectionString) == 0 || len(queueName) == 0)) {
		fmt.Println("sb-shovel v0.1\nManage large message operations on a given Service Bus.\n ")
		fmt.Println("Example Usage:\n\tsb-shovel.exe -cmd dump -conn \"<servicebus_uri>\" -q queueName\n\tsb-shovel.exe -cmd emptyAll -conn \"<servicebus_uri>\" -q queueName -dlq -rq")
		flag.PrintDefaults()
		return
	}

	switch command {
	case commandList["dump"]:
		dump()
		return
	case commandList["emptyOne"]:
		empty(false, requeue)
		return
	case commandList["emptyAll"]:
		isEmpty = true
		empty(true, requeue)
		return
	case commandList["sendFromFile"]:
		if len(dir) == 0 {
			fmt.Println("Value for -dir flag missing")
			return
		}
		sendFromFile(dir)
		return
	}
}

func dump() {
	t := template.Must(template.New("output").Parse(tmpl))

	sb, err := setupServiceBus(connectionString)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	q, err := setupQueue(ctx, sb, queueName, isDlq, isEmpty)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	defer q.Close(ctx)

	totalMsgs, err := getMessageCount(ctx, sb, queueName, isDlq)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	fmt.Printf("%d messages to process on %s queue...\n", totalMsgs, q.Name)

	err = createDir()
	if err != nil {
		fmt.Println(err)
		return
	}

	// start work
	writeChannel := make(chan []string, 3)
	errChannel := make(chan string)
	go readQueue(writeChannel, errChannel, context.Background(), q, maxWriteCache, t)

	start := time.Now()
	var wg sync.WaitGroup
	i := 1
	fileNamePattern := map[int]string{1: "00000", 2: "0000", 3: "000", 4: "00", 5: "0"}

	done := false
	for !done {
		select {
		case msgs, ok := <-writeChannel:
			if !ok {
				done = true
				continue
			}
			wg.Add(1)
			go writeFile(errChannel, i, msgs, fileNamePattern[len(fmt.Sprint(i))], &wg)
			i++
		case e, _ := <-errChannel:
			fmt.Println(e)
			wg.Wait()
			return
		}
	}
	fmt.Println("All messages have been read, awaiting file writes...")
	wg.Wait()
	fmt.Printf("Finished in %dms\n", time.Since(start).Milliseconds())
}

func empty(all, requeue bool) {
	sb, err := setupServiceBus(connectionString)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	sourceQ, err := setupQueue(ctx, sb, queueName, isDlq, isEmpty)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	defer sourceQ.Close(ctx)

	totalMsgs, err := getMessageCount(ctx, sb, queueName, isDlq)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	fmt.Printf("%d messages to process on %s queue...\n", totalMsgs, sourceQ.Name)
	var start time.Time

	if totalMsgs > 0 {
		if requeue {
			targetQ, err := setupQueue(ctx, sb, queueName, !isDlq, false)
			if err != nil {
				fmt.Println("ERROR: ", err)
				return
			}
			defer targetQ.Close(ctx)
			start = time.Now()
			emptyQueue(all, requeue, context.Background(), sourceQ, targetQ, totalMsgs)
		} else {
			start = time.Now()
			emptyQueue(all, requeue, context.Background(), sourceQ, nil, totalMsgs)
		}
		fmt.Printf("Finished %d messages in %dms\n", totalMsgs, time.Since(start).Milliseconds())
	}
}

func sendFromFile(dir string) {
	sb, err := setupServiceBus(connectionString)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	q, err := setupQueue(ctx, sb, queueName, isDlq, false)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	defer q.Close(ctx)

	data := readFile(dir)
	if len(data) == 0 {
		fmt.Println("Nothing to send")
		return
	}

	fmt.Printf("Sending %d messages\n", len(data))

	for i := 0; i < len(data); i++ {
		msg := &servicebus.Message{
			Data:        data[i],
			ContentType: "application/json",
		}
		err := q.Send(ctx, msg)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Finished sending messages")
}
