package cmd

import (
	"errors"
	"fmt"

	sbio "github.com/aagoldingay/sb-shovel/io"
	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
	"github.com/spf13/cobra"
)

var file string

var sendCmd = &cobra.Command{
	Use:     "send",
	Example: "sb-shovel send -c <connection_string> -q <queuename> -f path/to/file.json",
	Short:   "send messages to a targeted queue",
	Long:    "WARNING: max read size for a file line is 64*4096 characters\nWARNING: ensure messages are properly formatted before sending",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(file) == 0 {
			return errors.New("value for -file flag is not valid")

		}
		err := sendFromFile(sb, queue, file)
		if err != nil {
			return err
		}
		return nil
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
