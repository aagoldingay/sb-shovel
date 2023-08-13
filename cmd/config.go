package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// parent command, no run
var configCmd = &cobra.Command{
	Use:     "config",
	Example: "sb-shovel config list\nsb-shovel config update EASY_IDENTIFIER <connection_string>\nsb-shovel config delete EASY_IDENTIFIER",
	Short:   "manage saved connections",
	Long:    `saved connections make running sb-shovel much easier, by using the custom identifier in the connection string parameter instead`,
	Args:    cobra.MinimumNArgs(1),
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "retrieve list of connections",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("key0 | value0")
		fmt.Println("key1 | value1")
		fmt.Println("key2 | value2")
	},
}

var configUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "add new or edit existing connection string identifiers",
	Long:  "accepts two values - human readable name (key) and the real connection string (value)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("saved %s to %s\n", args[1], args[0])
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "remove a saved connection string identifier",
	Long:  "accepts one value - human readable name (key). use 'list' to identify stored key value pairs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("removed %s\n", args[0])
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUpdateCmd)
	configCmd.AddCommand(configDeleteCmd)
	rootCmd.AddCommand(configCmd)
}
