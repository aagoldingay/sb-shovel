package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var delay bool

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Example: "sb-shovel delete -c <connection_string> -q <queuename>\nsb-shovel  delete -c <connection_string> -q <queuename> --dlq --all",
	Short:   "delete messages from a targeted queue",
	Long:    "WARNING: execution without '-delay' may cause issues if you are dealing with extremely large queues",
	Run: func(cmd *cobra.Command, args []string) {
		s := "one"
		if all {
			s = "all"
		}
		fmt.Printf("deleting %s message(s) from %s queue \n", s, queue)
		if delay {
			fmt.Println("250ms delay between message batches")
		}
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
