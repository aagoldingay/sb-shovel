package cmd

import (
	"fmt"

	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
	"github.com/spf13/cobra"
)

var requeueCmd = &cobra.Command{
	Use:     "requeue",
	Example: "sb-shovel requeue -c <connection_string> -q <queuename>\nsb-shovel  requeue -c <connection_string> -q <queuename> --dlq --all",
	Short:   "requeue messages from deadletter to the corresponding active queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := requeue(sb, queue, all, is_dlq)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	requeueCmd.Flags().StringVarP(&queue, "queue", "q", "", "name of the service bus queue")
	requeueCmd.Flags().StringVarP(&conn, "connection-string", "c", "", "connection string of the service bus resource")
	requeueCmd.Flags().BoolVar(&is_dlq, "dlq", false, "target the deadletter subqueue")
	requeueCmd.Flags().BoolVar(&all, "all", false, "process all messages on the queue")
	requeueCmd.MarkFlagRequired("queue")
	requeueCmd.MarkFlagRequired("connection-string")
	rootCmd.AddCommand(requeueCmd)
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
