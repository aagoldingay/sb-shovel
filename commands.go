package main

import (
	"fmt"
	"sync"
	"time"

	sbio "sb-shovel/io"
	sbc "sb-shovel/sbcontroller"
)

func doMain(connectionString string) {
	sb, err := sbc.NewController(connectionString)
	if err != nil {
		return
	}
	var cmds = map[string]string{"dump": "dump", "emptyOne": "emptyOne", "emptyAll": "emptyAll", "sendFromFile": "sendFromFile"}
	command := "dump"
	requeue := false
	dir := ""
	dlq := true
	maxWrite := 5
	switch command {
	case cmds["dump"]:
		doDump(sb, "queuename", dlq, maxWrite)
		return
	case cmds["emptyOne"]:
		doEmpty(sb, "queuename", dlq, false, requeue)
		return
	case cmds["emptyAll"]:
		doEmpty(sb, "queuename", dlq, true, requeue)
		return
	case cmds["sendFromFile"]:
		if len(dir) == 0 {
			fmt.Println("Value for -dir flag missing")
			return
		}
		doSendFromFile(sb, "queuename", dir)
		return
	}
}

func doDump(sb sbc.Controller, q string, dlq bool, maxWrite int) error {
	err := sb.SetupSourceQueue(q, dlq, false)
	if err != nil {
		return err
	}
	defer sb.DisconnectSource()

	total, err := sb.GetSourceQueueCount()
	if err != nil {
		return err
	}
	if total == 0 {
		return fmt.Errorf("no messages on queue")
	}

	fmt.Printf("%d messages to process on %s queue...\n", total, q)

	// create dir
	err = sbio.CreateDir()
	if err != nil {
		return err
	}

	// start work
	returnedMsgs := make(chan []string)
	eChan := make(chan error)

	start := time.Now()
	go sb.ReadSourceQueue(returnedMsgs, eChan, maxWrite)

	var wg sync.WaitGroup

	done := false
	fileCount := 1
	for !done {
		select {
		case msgs, ok := <-returnedMsgs:
			if !ok {
				done = true
				continue
			}
			wg.Add(1)
			go sbio.WriteFile(eChan, fileCount, msgs, &wg)
			fileCount++
		case e := <-eChan:
			if e.Error() == sbc.ERR_QUEUEEMPTY {
				close(returnedMsgs)
				wg.Wait()
				break
			}
			wg.Wait()
			return e
		}
	}
	wg.Wait()
	close(eChan)
	fmt.Printf("Finished in %dms\n", time.Since(start).Milliseconds())
	return nil
}

func doEmpty(sb sbc.Controller, q string, dlq, all, requeue bool) error {
	err := sb.SetupSourceQueue(q, dlq, true)
	if err != nil {
		return err
	}
	defer sb.DisconnectSource()

	if requeue {
		if !dlq == true {
			return fmt.Errorf("cannot requeue messages directly to a dead letter queue")
		}
		err = sb.SetupTargetQueue(q, !dlq, true)
		if err != nil {
			return fmt.Errorf("problem setting up target queue: %v", err)
		}
	}

	c, err := sb.GetSourceQueueCount()
	if err != nil {
		return err
	}
	if c == 0 {
		return fmt.Errorf("no messages to delete")
	}

	if all {
		eChan := make(chan error)
		go sb.DeleteManyMessages(eChan, requeue, c)

		done := false
		for !done {
			e := <-eChan
			if e.Error() == sbc.ERR_DELETESTATUS {
				fmt.Println(e.Error())
			}
			if e.Error() != "context canceled" {
				return err
			}
			done = true
		}
		close(eChan)
	} else {
		err = sb.DeleteOneMessage(requeue)
		if err != nil {
			return err
		}
	}

	curr, err := sb.GetSourceQueueCount()
	if err != nil {
		return err
	}
	completeMessage := "%d deleted"
	if curr > 0 {
		completeMessage += fmt.Sprintf(", %d remaining - these may have been added since the process began", curr)
	}
	fmt.Printf(fmt.Sprintf("%s\n", completeMessage), c)

	if requeue {
		sb.DisconnectTarget()
	}

	return nil
}

func doSendFromFile(sb sbc.Controller, q, dir string) error {
	err := sb.SetupSourceQueue(q, false, true)
	if err != nil {
		return err
	}
	defer sb.DisconnectSource()

	// read from file
	data := sbio.ReadFile(dir)

	err = sb.SendManyJsonMessages(false, data)
	if err != nil {
		return err
	}
	return nil
}
