package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var requeueCmd = &cobra.Command{
	Use:     "requeue",
	Example: "sb-shovel requeue -c <connection_string> -q <queuename>\nsb-shovel  requeue -c <connection_string> -q <queuename> --dlq --all",
	Short:   "requeue messages from deadletter to the corresponding active queue",
	Run: func(cmd *cobra.Command, args []string) {
		s := "one"
		if all {
			s = "all"
		}
		fmt.Printf("deleting %s message(s) from %s queue \n", s, queue)
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
