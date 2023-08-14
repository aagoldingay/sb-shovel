package cmd

import (
	"fmt"
	"strings"

	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
	"github.com/spf13/cobra"
)

var delay bool

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Example: "sb-shovel delete -c <connection_string> -q <queuename>\nsb-shovel  delete -c <connection_string> -q <queuename> --dlq --all",
	Short:   "delete messages from a targeted queue",
	Long:    "WARNING: execution without '-delay' may cause issues if you are dealing with extremely large queues",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := delete(sb, queue, is_dlq, all, delay)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&queue, "queue", "q", "", "name of the service bus queue")
	deleteCmd.Flags().StringVarP(&conn, "connection-string", "c", "", "connection string of the service bus resource")
	deleteCmd.Flags().BoolVar(&all, "all", false, "process all messages on the queue")
	deleteCmd.Flags().BoolVar(&is_dlq, "dlq", false, "target the deadletter subqueue")
	deleteCmd.Flags().BoolVar(&delay, "delay", false, "enforces a 250ms delay for every 50 messages processed")
	deleteCmd.MarkFlagRequired("queue")
	deleteCmd.MarkFlagRequired("connection-string")
	rootCmd.AddCommand(deleteCmd)
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
