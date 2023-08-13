package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// many commands share variables - those are declared here
var queue, conn string
var debug, is_dlq, all bool

var rootCmd = &cobra.Command{
	Use:   "sb-shovel",
	Short: "manage large message operations on a given Service Bus",
	// Long:  ``,
	// persistent prerun to load config?
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "verbose logging for the executed command")
}

func initConfig() {
	// load config file
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
