package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	cc "github.com/aagoldingay/sb-shovel/config"
	sbio "github.com/aagoldingay/sb-shovel/io"
	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
)

func config(config cc.ConfigManager, args []string) {
	err := config.LoadConfig()
	if err != nil && err.Error() != cc.ERR_NOCONFIG {
		fmt.Println(err)
		return
	}

	switch args[0] {
	case "list":
		if len(args) != 1 {
			fmt.Println("unexpected arguments for list command\nusage: sb-shovel -cmd config list")
			return
		}
		fmt.Println(config.ListConfig())
	case "remove":
		if len(args) != 2 {
			fmt.Println("unexpected arguments for remove command\nusage: sb-shovel -cmd config remove KEY_NAME")
			return
		}
		err := config.DeleteConfigValue(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		if err = config.SaveConfig(); err != nil {
			fmt.Printf("error saving config: %s\n", err.Error())
			return
		}
		fmt.Printf("%s removed from config\n", args[1])
	case "update":
		if len(args) != 3 {
			fmt.Println("unexpected arguments for update command\nusage: sb-shovel -cmd config update KEY_NAME KEY_VALUE")
			return
		}
		config.UpdateConfig(args[1], args[2])
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("error saving config: %s\n", err.Error())
			return
		}

		fmt.Printf("%s value updated in config\n", args[1])
	default:
		fmt.Println("unexpected config command provided")
	}
}

func pull(sb sbc.Controller, q string, dlq bool, maxWrite int) error {
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

func requeue(sb sbc.Controller, q string, all, dlq bool) error {
	if !dlq {
		return fmt.Errorf("cannot requeue messages directly to a dead letter queue")
	}

	err := sb.SetupSourceQueue(q, dlq, true)

	if err != nil {
		return err
	}

	err = sb.SetupTargetQueue(q, !dlq, true)
	if err != nil {
		return fmt.Errorf("problem setting up target queue: %v", err)
	}
	defer sb.DisconnectQueues()

	if all {
		c, err := sb.GetSourceQueueCount()

		if err != nil {
			return err
		}

		if c == 0 {
			return fmt.Errorf("no messages to requeue")
		}

		fmt.Printf("%d messages to requeue\n", c)
		err = sb.RequeueManyMessages(c)
		if err != nil {
			return err
		}
		fmt.Println("messages requeued")
	} else {
		err = sb.RequeueOneMessage()
		if err != nil {
			return err
		}
		fmt.Println("one message requeued")
	}

	return nil
}

func delete(sb sbc.Controller, q string, dlq, all, delay bool) error {
	err := sb.SetupSourceQueue(q, dlq, true)

	if err != nil {
		return err
	}
	defer sb.DisconnectSource()

	c, err := sb.GetSourceQueueCount()

	if err != nil {
		return err
	}
	if c == 0 {
		return fmt.Errorf("no messages to delete")
	}

	if all {
		fmt.Printf("%d messages to delete\n", c)
		eChan := make(chan error)
		go sb.DeleteManyMessages(eChan, c, delay)

		done := false
		for !done {
			e := <-eChan
			if strings.Contains(e.Error(), "[status]") {
				fmt.Print(e.Error())
				continue
			}
			if e.Error() != "context canceled" {
				return e
			}
			fmt.Println("") // blank line to separate status overwrites from completion messages
			done = true
		}
		close(eChan)
	} else {
		err = sb.DeleteOneMessage()
		if err != nil {
			return err
		}
		fmt.Println("1 message deleted")
		return nil
	}

	fmt.Printf("%d message(s) deleted\n", c)

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
