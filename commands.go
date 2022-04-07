package main

import (
	"fmt"
	"sync"
	"time"

	sbio "sb-shovel/io"
	sbc "sb-shovel/sbcontroller"
)

func dump(sb sbc.Controller, q string, dlq bool, maxWrite int) error {
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

	err = sbio.CreateDir()
	if err != nil {
		return err
	}

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

func empty(sb sbc.Controller, q string, dlq, all, requeue bool) error {
	err := sb.SetupSourceQueue(q, dlq, true)

	if err != nil {
		return err
	}
	defer sb.DisconnectSource()

	if requeue {
		if !dlq {
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
		msgCount := "%d messages to %s\n"
		if requeue {
			fmt.Printf(msgCount, c, "requeue")
		} else {
			fmt.Printf(msgCount, c, "delete")
		}
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
		if requeue {
			fmt.Println("1 message requeued")
		} else {
			fmt.Println("1 message deleted")
		}
		if requeue {
			sb.DisconnectTarget()
		}
		return nil
	}

	curr, err := sb.GetSourceQueueCount()

	if err != nil {
		return err
	}

	// completeMessage := "%d %s"
	completeMessage := "%d message(s) %s"

	if requeue {
		completeMessage = fmt.Sprintf(completeMessage, c, "requeued")
	} else {
		completeMessage = fmt.Sprintf(completeMessage, c, "deleted")
	}

	if curr > 0 {
		completeMessage += fmt.Sprintf(", %d remaining - these may have been added since the process began", curr)
	}

	fmt.Printf("%s\n", completeMessage)

	if requeue {
		sb.DisconnectTarget()
	}

	return nil
}

func sendFromFile(sb sbc.Controller, q, dir string) error {
	err := sb.SetupSourceQueue(q, false, true)
	if err != nil {
		return err
	}
	defer sb.DisconnectSource()

	data := sbio.ReadFile(dir)

	err = sb.SendManyJsonMessages(false, data)
	if err != nil {
		return err
	}
	fmt.Printf("Sent %d messages\n", len(data))
	return nil
}
