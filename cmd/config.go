package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// parent command, no run
var configCmd = &cobra.Command{
	Use:     "config",
	Example: "sb-shovel config list\nsb-shovel config update EASY_IDENTIFIER <connection_string>\nsb-shovel config remove EASY_IDENTIFIER",
	Short:   "manage saved connections",
	Long:    `saved connections make running sb-shovel much easier, by using the custom identifier in the connection string parameter instead`,
	Args:    cobra.MinimumNArgs(1),
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "retrieve list of connections",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		fmt.Println(cfg.ListConfig())
	},
}

var configUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "add new or edit existing connection string identifiers",
	Long:  "accepts two values - human readable name (key) and the real connection string (value)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		cfg.UpdateConfig(args[0], args[1])
		if err := cfg.SaveConfig(); err != nil {
			fmt.Printf("error saving config: %s\n", err.Error())
			return
		}

		fmt.Printf("%s value updated in config\n", args[0])
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a saved connection string identifier",
	Long:  "accepts one value - human readable name (key). use 'list' to identify stored key value pairs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		err := cfg.DeleteConfigValue(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		if err = cfg.SaveConfig(); err != nil {
			fmt.Printf("error saving config: %s\n", err.Error())
			return
		}
		fmt.Printf("%s removed from config\n", args[0])
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUpdateCmd)
	configCmd.AddCommand(configRemoveCmd)
	rootCmd.AddCommand(configCmd)
}
