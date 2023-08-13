package cmd

import (
	"fmt"

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
	Run: func(cmd *cobra.Command, args []string) {
		s := fmt.Sprintf("pulling messages from %s", queue)
		if is_dlq {
			fmt.Printf("%s/deadletter\n", s)
		} else {
			fmt.Printf("%s\n", s)
		}
		fmt.Printf("up to %d messages per file\n", n_lines)
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
