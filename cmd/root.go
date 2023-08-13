package cmd

import (
	"fmt"
	"os"
	"strings"

	cc "github.com/aagoldingay/sb-shovel/config"
	sbc "github.com/aagoldingay/sb-shovel/sbcontroller"
	"github.com/spf13/cobra"
)

// many commands share variables - those are declared here
var queue, conn string
var debug, is_dlq, all bool
var cfg cc.ConfigManager
var sb sbc.Controller

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
	var err error
	cfg, err = cc.NewConfigController("sb-shovel")
	if err != nil {
		fmt.Println(err)
	}
	err = cfg.LoadConfig()
	if err != nil && err.Error() != cc.ERR_NOCONFIG {
		fmt.Println(err)
		return
	}
}

func initServiceBus(connectionString string) {
	var err error
	sb, err = sbc.NewServiceBusController(connectionString)
	if err != nil {
		fmt.Println(err)
	}
}

func checkIfConfig(s string) (bool, string) {
	if strings.HasPrefix(s, "cfg|") {
		return true, strings.Split(s, "|")[1]
	}
	return false, ""
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
