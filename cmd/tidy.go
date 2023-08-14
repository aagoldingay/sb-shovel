package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
	"github.com/spf13/cobra"
)

var pattern string
var execute bool

var tidyCmd = &cobra.Command{
	Use:     "tidy",
	Example: "sb-shovel tidy -c <connection_string> -q <queuename> -pattern \"[A-Z]{3}\"\nsb-shovel tidy -c <connection_string> -q <queuename> -pattern \"[A-Z]{3}\" -x",
	Short:   "selectively delete messages containing a regex pattern",
	Long:    "without passing --execute, this command will only peek and pattern match\nWARNING: using this command abandons messages that are not matched.\nNOTE: refer to the approved syntax: https://github.com/google/re2/wiki/Syntax",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(pattern) == 0 {
			return errors.New("pattern must be specified, else all messages risk being deleted")
		}
		err := tidy(sb, queue, pattern, is_dlq, execute)
		if err != nil {
			return err
		}
		return nil
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

func tidy(sb sbc.Controller, q, pattern string, dlq, execute bool) error {
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
		return fmt.Errorf("no messages to process")
	}

	rex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Problem compiling regex. Refer to the approved syntax: https://github.com/google/re2/wiki/Syntax")
		return err
	}

	if !execute {
		fmt.Println("Tidy executing as a dry run. Pass '-x' to action")
	}

	eChan := make(chan error)
	go sb.TidyMessages(eChan, rex, execute, c)

	done := false
	for !done {
		e := <-eChan
		if strings.Contains(e.Error(), "[status]") {
			fmt.Printf("%s\n", e.Error())
			continue
		}
		if e.Error() != "context canceled" && e.Error() != sbc.ERR_QUEUEEMPTY {
			return e
		}
		done = true
	}
	close(eChan)

	return nil
}
