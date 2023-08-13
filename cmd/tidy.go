package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pattern string
var execute bool

var tidyCmd = &cobra.Command{
	Use:     "tidy",
	Example: "sb-shovel tidy -c <connection_string> -q <queuename> -pattern \"[A-Z]{3}\"\nsb-shovel tidy -c <connection_string> -q <queuename> -pattern \"[A-Z]{3}\" -x",
	Short:   "selectively delete messages containing a regex pattern",
	Long:    "without passing --execute, this command will only peek and pattern match\nWARNING: using this command abandons messages that are not matched.\nNOTE: refer to the approved syntax: https://github.com/google/re2/wiki/Syntax",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tidying messages from %s queue matching %s\n", queue, pattern)
		if !execute {
			fmt.Println("pass --execute with the command to delete")
		}
	},
}

func init() {
	tidyCmd.Flags().StringVarP(&queue, "queue", "q", "", "name of the service bus queue")
	tidyCmd.Flags().StringVarP(&conn, "connection-string", "c", "", "connection string of the service bus resource")
	tidyCmd.Flags().StringVarP(&pattern, "pattern", "p", "", "regex pattern to match against message contents")
	tidyCmd.Flags().BoolVar(&is_dlq, "dlq", false, "target the deadletter subqueue")
	tidyCmd.Flags().BoolVarP(&execute, "execute", "x", false, "run the command")
	tidyCmd.MarkFlagRequired("queue")
	tidyCmd.MarkFlagRequired("connection-string")
	tidyCmd.MarkFlagRequired("pattern")
	rootCmd.AddCommand(tidyCmd)
}
