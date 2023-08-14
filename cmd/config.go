package cmd

import (
	"fmt"

	cc "github.com/aagoldingay/sb-shovel/config"
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
		config(cfg, append([]string{"list"}, args...))
	},
}

var configUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "add new or edit existing connection string identifiers",
	Long:  "accepts two values - human readable name (key) and the real connection string (value)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		config(cfg, append([]string{"update"}, args...))
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a saved connection string identifier",
	Long:  "accepts one value - human readable name (key). use 'list' to identify stored key value pairs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		config(cfg, append([]string{"remove"}, args...))
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUpdateCmd)
	configCmd.AddCommand(configRemoveCmd)
	rootCmd.AddCommand(configCmd)
}

func config(config cc.ConfigManager, args []string) {
	err := config.LoadConfig()
	if err != nil && err.Error() != cc.ERR_NOCONFIG {
		fmt.Println(err)
		return
	}

	switch args[0] {
	case "list":
		if len(args) != 1 {
			fmt.Println("unexpected arguments for list command\nusage: sb-shovel -cmd config list")
			return
		}
		fmt.Println(config.ListConfig())
	case "remove":
		if len(args) != 2 {
			fmt.Println("unexpected arguments for remove command\nusage: sb-shovel -cmd config remove KEY_NAME")
			return
		}
		err := config.DeleteConfigValue(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		if err = config.SaveConfig(); err != nil {
			fmt.Printf("error saving config: %s\n", err.Error())
			return
		}
		fmt.Printf("%s removed from config\n", args[1])
	case "update":
		if len(args) != 3 {
			fmt.Println("unexpected arguments for update command\nusage: sb-shovel -cmd config update KEY_NAME KEY_VALUE")
			return
		}
		config.UpdateConfig(args[1], args[2])
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("error saving config: %s\n", err.Error())
			return
		}

		fmt.Printf("%s value updated in config\n", args[1])
	default:
		fmt.Println("unexpected config command provided")
	}
}
