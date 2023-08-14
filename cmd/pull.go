package cmd

import (
	"errors"
	"fmt"
	"sync"
	"time"

	sbio "github.com/aagoldingay/sb-shovel/io"
	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
	"github.com/spf13/cobra"
)

// var template string
var n_lines int

// parent command, no run
var pullCmd = &cobra.Command{
	Use:     "pull",
	Example: "sb-shovel pull -c <connection_string> -q <queuename>\nsb-shovel pull -c <connection_string> -q <queuename> --dlq",
	Short:   "manage saved connections",
	Long:    "WARNING: local files with the same naming pattern will be overwritten",
	RunE: func(cmd *cobra.Command, args []string) error {
		if n_lines < 1 {
			return errors.New("Value for -n-lines is not valid. Must be >= 1")
		}

		err := pull(sb, queue, is_dlq, n_lines)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	pullCmd.Flags().StringVarP(&queue, "queue", "q", "", "name of the service bus queue")
	pullCmd.Flags().StringVarP(&conn, "connection-string", "c", "", "connection string of the service bus resource")
	pullCmd.Flags().IntVarP(&n_lines, "n-lines", "n", 100, "number of lines per file")
	pullCmd.Flags().BoolVar(&is_dlq, "dlq", false, "target the deadletter subqueue")
	pullCmd.MarkFlagRequired("queue")
	pullCmd.MarkFlagRequired("connection-string")
	rootCmd.AddCommand(pullCmd)
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
