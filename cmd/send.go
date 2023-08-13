package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var file string

var sendCmd = &cobra.Command{
	Use:     "send",
	Example: "sb-shovel send -c <connection_string> -q <queuename> -f path/to/file.json",
	Short:   "send messages to a targeted queue",
	Long:    "WARNING: max read size for a file line is 64*4096 characters\nWARNING: ensure messages are properly formatted before sending",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sending messages to %s queue from %s\n", queue, file)
	},
}

func init() {
	sendCmd.Flags().StringVarP(&queue, "queue", "q", "", "name of the service bus queue")
	sendCmd.Flags().StringVarP(&conn, "connection-string", "c", "", "connection string of the service bus resource")
	sendCmd.Flags().StringVarP(&file, "file", "f", "", "path to file containing messages to send")
	sendCmd.MarkFlagRequired("queue")
	sendCmd.MarkFlagRequired("connection-string")
	sendCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(sendCmd)
}
